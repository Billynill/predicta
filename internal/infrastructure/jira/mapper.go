package jira

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/predicta/predicta/internal/domain/entity"
)

const sprintIssueFields = "summary,status,assignee"

func mapIssueToTask(issue issueRecord, sprintID string) entity.Task {
	assigneeID := ""
	if issue.Fields.Assignee != nil {
		assigneeID = issue.Fields.Assignee.AccountID
	}

	points := 1
	if issue.Fields.Points != nil && *issue.Fields.Points > 0 {
		points = int(math.Round(*issue.Fields.Points))
	}

	return entity.Task{
		ID:          issue.Key,
		ExternalID:  issue.Key,
		Title:       issue.Fields.Summary,
		AssigneeID:  assigneeID,
		Status:      mapStatus(issue.Fields.Status),
		StoryPoints: points,
		SprintID:    sprintID,
	}
}

func mapStatus(st issueStatus) entity.TaskStatus {
	switch st.StatusCategory.Key {
	case "done":
		return entity.TaskStatusDone
	case "new":
		return entity.TaskStatusTodo
	default:
		return entity.TaskStatusInProgress
	}
}

func sprintDaysRemaining(endDate string, fallback int) int {
	end, err := parseJiraTime(endDate)
	if err != nil || end.IsZero() {
		if fallback > 0 {
			return fallback
		}
		return 1
	}
	now := time.Now().UTC()
	if !end.After(now) {
		return 0
	}
	days := int(math.Ceil(end.Sub(now).Hours() / 24))
	if days < 1 {
		return 1
	}
	return days
}

func extractStoryPoints(raw map[string]any) *float64 {
	keys := []string{
		"customfield_10016",
		"customfield_10020",
		"customfield_10026",
		"story_points",
	}
	for _, k := range keys {
		if v, ok := raw[k]; ok {
			if f, ok := toFloat(v); ok {
				return &f
			}
		}
	}
	return nil
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

func sprintIDString(id int) string {
	return strconv.Itoa(id)
}

func mapAssignable(u jiraUser) AssignableUser {
	return AssignableUser{
		AccountID:   u.AccountID,
		DisplayName: u.DisplayName,
		Email:       u.EmailAddress,
	}
}

func fmtSprintName(s sprintResponse) string {
	if s.Name != "" {
		return s.Name
	}
	return fmt.Sprintf("Sprint %d", s.ID)
}
