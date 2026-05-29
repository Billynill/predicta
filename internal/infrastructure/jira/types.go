package jira

import "time"

type sprintResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	State     string `json:"state"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Goal      string `json:"goal"`
}

type sprintListResponse struct {
	Values []sprintResponse `json:"values"`
}

type boardListResponse struct {
	Values []struct {
		ID int `json:"id"`
	} `json:"values"`
}

type issuesPageResponse struct {
	StartAt    int           `json:"startAt"`
	MaxResults int           `json:"maxResults"`
	Total      int           `json:"total"`
	Issues     []issueRecord `json:"issues"`
}

type issueRecord struct {
	ID     string      `json:"id"`
	Key    string      `json:"key"`
	Fields issueFields `json:"fields"`
}

type issueFields struct {
	Summary  string          `json:"summary"`
	Assignee *jiraUser       `json:"assignee"`
	Status   issueStatus     `json:"status"`
	Points   *float64        `json:"-"` // filled from custom fields when present
	Raw      map[string]any  `json:"-"`
}

type issueStatus struct {
	Name           string `json:"name"`
	StatusCategory struct {
		Key string `json:"key"`
	} `json:"statusCategory"`
}

type jiraUser struct {
	AccountID    string            `json:"accountId"`
	DisplayName  string            `json:"displayName"`
	EmailAddress string            `json:"emailAddress"`
	AvatarURLs   map[string]string `json:"avatarUrls"`
}

type usersSearchResponse struct {
	Total      int        `json:"total"`
	StartAt    int        `json:"startAt"`
	MaxResults int        `json:"maxResults"`
	Values     []jiraUser `json:"values"`
}

// AssignableUser — для настройки config/employees.json.
type AssignableUser struct {
	AccountID   string `json:"account_id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// ProjectSummary — проект Jira (для выбора JIRA_PROJECT_KEY).
type ProjectSummary struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

func parseJiraTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	// Jira: 2024-01-15T10:00:00.000Z
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000Z0700",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, nil
}
