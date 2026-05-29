package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/predicta/predicta/internal/delivery/http/dto"
	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/usecase"
)

type Handler struct {
	project  usecase.ProjectStatusGetter
	team     usecase.TeamVelocityGetter
	insights usecase.TeamInsightsGetter
	employee usecase.EmployeeAnalyticsGetter
	profile  usecase.EmployeeProfileGetter
	task     usecase.TaskReassigner
	creator  usecase.TaskCreator
}

func New(
	project usecase.ProjectStatusGetter,
	team usecase.TeamVelocityGetter,
	insights usecase.TeamInsightsGetter,
	employee usecase.EmployeeAnalyticsGetter,
	profile usecase.EmployeeProfileGetter,
	task usecase.TaskReassigner,
	creator usecase.TaskCreator,
) *Handler {
	return &Handler{
		project:  project,
		team:     team,
		insights: insights,
		employee: employee,
		profile:  profile,
		task:     task,
		creator:  creator,
	}
}

func (h *Handler) Register(r *gin.RouterGroup) {
	api := r.Group("/api")
	{
		api.GET("/project/status", h.getProjectStatus)
		api.GET("/team/velocity", h.getTeamVelocity)
		api.GET("/team/insights", h.getTeamInsights)
		api.GET("/employee/:id", h.getEmployeeProfile)
		api.GET("/employee/:id/analytics", h.getEmployeeAnalytics)
		api.POST("/tasks/create", h.createTask)
		api.POST("/tasks/reassign", h.reassignTask)
	}
}

func (h *Handler) getProjectStatus(c *gin.Context) {
	status, err := h.project.GetStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, mapProjectStatus(status))
}

func (h *Handler) getTeamVelocity(c *gin.Context) {
	items, err := h.team.GetVelocity(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	resp := make([]dto.EmployeeVelocityResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, dto.EmployeeVelocityResponse{
			ID:         item.Employee.ID,
			Name:       item.Employee.Name,
			Role:       item.Employee.Role,
			DoneCount:  item.DoneCount,
			TotalCount: item.TotalCount,
			Health:     string(item.Health),
		})
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) getTeamInsights(c *gin.Context) {
	insight, err := h.insights.GetInsights(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.TeamInsightsResponse{AIInsight: insight})
}

func (h *Handler) getEmployeeProfile(c *gin.Context) {
	id := c.Param("id")
	profile, err := h.profile.GetProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, mapEmployeeProfile(profile))
}

func (h *Handler) getEmployeeAnalytics(c *gin.Context) {
	id := c.Param("id")
	analytics, err := h.employee.GetAnalytics(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	tasks := make([]dto.TaskResponse, 0, len(analytics.Tasks))
	for _, t := range analytics.Tasks {
		tasks = append(tasks, dto.TaskResponse{
			ID:     t.ID,
			Title:  t.Title,
			Status: string(t.Status),
		})
	}

	c.JSON(http.StatusOK, dto.EmployeeAnalyticsResponse{
		EmployeeID:     analytics.Employee.ID,
		EmployeeName:   analytics.Employee.Name,
		Role:           analytics.Employee.Role,
		ForecastDays:   analytics.Forecast.DaysToComplete,
		SprintDaysLeft: analytics.Forecast.SprintDaysLeft,
		DelayDays:      analytics.Forecast.DelayDays,
		AIInsight:      analytics.AIInsight,
		Tasks:          tasks,
	})
}

func (h *Handler) createTask(c *gin.Context) {
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.creator.Create(c.Request.Context(), usecase.CreateTaskCommand{
		Title:       req.Title,
		Description: req.Description,
		AssigneeID:  req.AssigneeID,
		Force:       req.Force,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapCreateTaskResult(result))
}

func (h *Handler) reassignTask(c *gin.Context) {
	var req dto.ReassignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.task.Reassign(c.Request.Context(), req.TaskID, req.NewExecutorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ReassignResponse{
		Message:       result.Message,
		ProjectStatus: mapProjectStatus(result.ProjectStatus),
	})
}

func mapEmployeeProfile(p entity.EmployeeProfile) dto.EmployeeProfileResponse {
	tasks := make([]dto.TaskResponse, 0, len(p.Tasks))
	for _, t := range p.Tasks {
		tasks = append(tasks, dto.TaskResponse{
			ID:     t.ID,
			Title:  t.Title,
			Status: string(t.Status),
		})
	}
	return dto.EmployeeProfileResponse{
		ID:           p.Employee.ID,
		Name:         p.Employee.Name,
		Role:         p.Employee.Role,
		TelegramNick: p.Employee.TelegramNick,
		AvatarURL:    p.AvatarURL,
		DoneCount:    p.DoneCount,
		TotalCount:   p.TotalCount,
		Remaining:    p.Remaining,
		Health:       string(p.Health),
		AIInsight:    p.AIInsight,
		Tasks:        tasks,
	}
}

func mapCreateTaskResult(r entity.CreateTaskResult) dto.CreateTaskResponse {
	resp := dto.CreateTaskResponse{
		Created:      r.Created,
		Approved:     r.Approved,
		TaskID:       r.TaskID,
		TaskTitle:    r.TaskTitle,
		AssigneeID:   r.AssigneeID,
		AssigneeName: r.AssigneeName,
		AIInsight:    r.AIInsight,
	}
	if !r.Approved {
		resp.SuggestedAssigneeID = r.SuggestedAssigneeID
		resp.SuggestedAssigneeName = r.SuggestedAssigneeName
	}
	return resp
}
