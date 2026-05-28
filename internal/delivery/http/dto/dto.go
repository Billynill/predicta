package dto

type ProjectStatusResponse struct {
	SprintName      string  `json:"sprint_name"`
	CompletionPct   float64 `json:"completion_pct"`
	DelayDays       int     `json:"delay_days"`
	IsAtRisk        bool    `json:"is_at_risk"`
	RiskMessage     string  `json:"risk_message"`
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

type TaskResponse struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

type EmployeeAnalyticsResponse struct {
	EmployeeID       string         `json:"employee_id"`
	EmployeeName     string         `json:"employee_name"`
	Role             string         `json:"role"`
	ForecastDays     int            `json:"forecast_days_to_complete"`
	SprintDaysLeft   int            `json:"sprint_days_left"`
	DelayDays        int            `json:"delay_days"`
	AIInsight        string         `json:"ai_insight"`
	Tasks            []TaskResponse `json:"tasks"`
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
