package entity

type VelocityHealth string

const (
	VelocityHealthGood VelocityHealth = "good"
	VelocityHealthBad  VelocityHealth = "bad"
)

type EmployeeVelocity struct {
	Employee   Employee
	DoneCount  int
	TotalCount int
	Health     VelocityHealth
}

type ProjectStatus struct {
	SprintName    string
	CompletionPct float64
	DelayDays     int
	IsAtRisk      bool
	RiskMessage   string
	TrackName     string
	DaysRemaining int
}

type EmployeeForecast struct {
	EmployeeID     string
	EmployeeName   string
	RemainingTasks int
	DaysToComplete int
	SprintDaysLeft int
	DelayDays      int
}

type EmployeeAnalytics struct {
	Employee  Employee
	Forecast  EmployeeForecast
	Tasks     []Task
	ChatLogs  []ChatMessage
	AIInsight string
}

type ReassignResult struct {
	TaskID          string
	TaskTitle       string
	NewAssigneeID   string
	NewAssigneeName string
	ProjectStatus   ProjectStatus
	Message         string
}
