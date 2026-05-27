package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/predicta/predicta/internal/delivery/http/dto"
	"github.com/predicta/predicta/internal/infrastructure/jira"
)

// JiraIntegration — вспомогательные ручки для настройки Jira (users, sprint, tasks).
type JiraIntegration struct {
	client *jira.Client
}

func NewJiraIntegration(client *jira.Client) *JiraIntegration {
	return &JiraIntegration{client: client}
}

func (h *JiraIntegration) Register(r *gin.Engine) {
	g := r.Group("/api/integrations/jira")
	g.GET("/me", h.whoAmI)
	g.GET("/users", h.listUsers)
	g.GET("/sprint", h.getSprint)
	g.GET("/tasks", h.listTasks)
}

func (h *JiraIntegration) whoAmI(c *gin.Context) {
	me, err := h.client.WhoAmI(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, me)
}

func (h *JiraIntegration) listUsers(c *gin.Context) {
	project := c.Query("project")
	users, err := h.client.ListAssignableUsers(c.Request.Context(), project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *JiraIntegration) getSprint(c *gin.Context) {
	sprint, err := h.client.GetSprintInfo(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":         sprint.ID,
		"name":       sprint.Name,
		"state":      sprint.State,
		"start_date": sprint.StartDate,
		"end_date":   sprint.EndDate,
	})
}

func (h *JiraIntegration) listTasks(c *gin.Context) {
	sprintID := c.Query("sprint_id")
	tasks, err := h.client.GetSprintTasks(c.Request.Context(), sprintID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}
