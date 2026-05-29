package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/predicta/predicta/internal/delivery/http/dto"
	"github.com/predicta/predicta/internal/domain/port"
)

// JiraIntegration — setup/dev ручки Jira (через port.JiraSetup).
type JiraIntegration struct {
	setup port.JiraSetup
}

func NewJiraIntegration(setup port.JiraSetup) *JiraIntegration {
	return &JiraIntegration{setup: setup}
}

func (h *JiraIntegration) Register(r *gin.RouterGroup) {
	g := r.Group("/api/integrations/jira")
	g.GET("/me", h.whoAmI)
	g.GET("/projects", h.listProjects)
	g.GET("/project-key", h.resolveProjectKey)
	g.GET("/users", h.listUsers)
	g.GET("/sprint", h.getSprint)
	g.GET("/tasks", h.listTasks)
}

func (h *JiraIntegration) whoAmI(c *gin.Context) {
	me, err := h.setup.WhoAmI(c.Request.Context())
	if err != nil {
		jiraError(c, err)
		return
	}
	c.JSON(http.StatusOK, me)
}

func (h *JiraIntegration) listProjects(c *gin.Context) {
	projects, err := h.setup.ListProjects(c.Request.Context())
	if err != nil {
		jiraError(c, err)
		return
	}
	c.JSON(http.StatusOK, projects)
}

func (h *JiraIntegration) resolveProjectKey(c *gin.Context) {
	key, err := h.setup.ResolveProjectKey(c.Request.Context())
	if err != nil {
		jiraError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"project_key": key})
}

func (h *JiraIntegration) listUsers(c *gin.Context) {
	project := c.Query("project")
	users, err := h.setup.ListAssignableUsers(c.Request.Context(), project)
	if err != nil {
		jiraError(c, err)
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *JiraIntegration) getSprint(c *gin.Context) {
	sprint, err := h.setup.GetSprintInfo(c.Request.Context())
	if err != nil {
		jiraError(c, err)
		return
	}
	c.JSON(http.StatusOK, sprint)
}

func (h *JiraIntegration) listTasks(c *gin.Context) {
	sprintID := c.Query("sprint_id")
	tasks, err := h.setup.GetSprintTasks(c.Request.Context(), sprintID)
	if err != nil {
		jiraError(c, err)
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func jiraError(c *gin.Context, err error) {
	status := http.StatusBadGateway
	msg := err.Error()
	if strings.Contains(msg, "No project with the provided key") ||
		strings.Contains(msg, "project key is required") ||
		strings.Contains(msg, "could not detect project key") {
		status = http.StatusBadRequest
	}
	c.JSON(status, dto.ErrorResponse{Error: msg})
}
