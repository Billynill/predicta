package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/predicta/predicta/internal/adapter/http/dto"
	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/usecase"
)

type Handler struct {
	project  *usecase.ProjectService
	team     *usecase.TeamService
	employee *usecase.EmployeeService
	task     *usecase.TaskService
}

func New(
	project *usecase.ProjectService,
	team *usecase.TeamService,
	employee *usecase.EmployeeService,
	task *usecase.TaskService,
) *Handler {
	return &Handler{
		project:  project,
		team:     team,
		employee: employee,
		task:     task,
	}
}

func (h *Handler) Register(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/project/status", h.getProjectStatus)
		api.GET("/team/velocity", h.getTeamVelocity)
		api.GET("/employee/:id/analytics", h.getEmployeeAnalytics)
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

func mapProjectStatus(s entity.ProjectStatus) dto.ProjectStatusResponse {
	return dto.ProjectStatusResponse{
		SprintName:    s.SprintName,
		CompletionPct: s.CompletionPct,
		DelayDays:     s.DelayDays,
		IsAtRisk:      s.IsAtRisk,
		RiskMessage:   s.RiskMessage,
		TrackName:     s.TrackName,
		DaysRemaining: s.DaysRemaining,
	}
}
