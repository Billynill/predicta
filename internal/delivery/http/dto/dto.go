package dto

type ProjectStatusResponse struct {
	SprintName      string  `json:"sprint_name"`
	CompletionPct   float64 `json:"completion_pct"`
	DelayDays       int     `json:"delay_days"`
	IsAtRisk        bool    `json:"is_at_risk"`
	RiskMessage     string  `json:"risk_message"`
	AIAdvice        string  `json:"ai_advice"`
	TrackName       string  `json:"track_name"`
	DaysRemaining   int     `json:"days_remaining"`
}

type EmployeeVelocityResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	DoneCount  int    `json:"done_count"`
	TotalCount int    `json:"total_count"`
	Health     string `json:"health"`
}

type TeamInsightsResponse struct {
	AIInsight string `json:"ai_insight"`
}

type TaskResponse struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

type EmployeeProfileResponse struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Role         string         `json:"role"`
	TelegramNick string         `json:"telegram_nick"`
	AvatarURL    string         `json:"avatar_url,omitempty"`
	DoneCount    int            `json:"done_count"`
	TotalCount   int            `json:"total_count"`
	Remaining    int            `json:"remaining_count"`
	Health       string         `json:"health"`
	AIInsight    string         `json:"ai_insight"`
	Tasks        []TaskResponse `json:"tasks"`
}

type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	AssigneeID  string `json:"assignee_id" binding:"required"`
	Force       bool   `json:"force"`
}

type CreateTaskResponse struct {
	Created               bool   `json:"created"`
	Approved              bool   `json:"approved"`
	TaskID                string `json:"task_id,omitempty"`
	TaskTitle             string `json:"task_title,omitempty"`
	AssigneeID            string `json:"assignee_id"`
	AssigneeName          string `json:"assignee_name"`
	SuggestedAssigneeID   string `json:"suggested_assignee_id,omitempty"`
	SuggestedAssigneeName string `json:"suggested_assignee_name,omitempty"`
	AIInsight             string `json:"ai_insight"`
}

type EmployeeAnalyticsResponse struct {
	EmployeeID     string         `json:"employee_id"`
	EmployeeName   string         `json:"employee_name"`
	Role           string         `json:"role"`
	ForecastDays   int            `json:"forecast_days_to_complete"`
	SprintDaysLeft int            `json:"sprint_days_left"`
	DelayDays      int            `json:"delay_days"`
	AIInsight      string         `json:"ai_insight"`
	Tasks          []TaskResponse `json:"tasks"`
}

type ReassignRequest struct {
	TaskID        string `json:"task_id" binding:"required"`
	NewExecutorID string `json:"new_executor_id" binding:"required"`
}

type ReassignResponse struct {
	Message       string                `json:"message"`
	ProjectStatus ProjectStatusResponse `json:"project_status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type RegisterManagerRequest struct {
	FirstName    string `json:"first_name" binding:"required"`
	LastName     string `json:"last_name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"`
	TelegramNick string `json:"telegram_nick" binding:"required"`
	Phone        string `json:"phone" binding:"required"`
	AvatarURL    string `json:"avatar_url"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type ManagerProfileResponse struct {
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Email             string `json:"email"`
	TelegramNick      string `json:"telegram_nick"`
	Phone             string `json:"phone"`
	SubordinatesCount int    `json:"subordinates_count"`
	AvatarURL         string `json:"avatar_url,omitempty"`
	JiraDisplayName   string `json:"jira_display_name,omitempty"`
	JiraEmail         string `json:"jira_email,omitempty"`
}
