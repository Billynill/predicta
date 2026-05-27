package entity

import "time"

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

type Employee struct {
	ID           string
	ExternalID   string // Jira accountId / Linear user id
	Name         string
	Role         string
	TelegramNick string
}

type Task struct {
	ID         string
	ExternalID string
	Title      string
	AssigneeID string
	Status     TaskStatus
	StoryPoints int
	SprintID   string
}

type Sprint struct {
	ID              string
	Name            string
	TeamID          string
	DaysRemaining   int
	TotalTasks      int
	CompletedTasks  int
}

type ChatMessage struct {
	ID         string
	EmployeeID string
	Text       string
	SentAt     time.Time
}

type EmployeeVelocity struct {
	Employee     Employee
	DoneCount    int
	TotalCount   int
	Health       VelocityHealth
}

type VelocityHealth string

const (
	VelocityHealthGood VelocityHealth = "good"
	VelocityHealthBad  VelocityHealth = "bad"
)

type ProjectStatus struct {
	SprintName       string
	CompletionPct    float64
	DelayDays        int
	IsAtRisk         bool
	RiskMessage      string
	TrackName        string
	DaysRemaining    int
}

type EmployeeForecast struct {
	EmployeeID       string
	EmployeeName     string
	RemainingTasks   int
	DaysToComplete   int
	SprintDaysLeft   int
	DelayDays        int
}

type EmployeeAnalytics struct {
	Employee   Employee
	Forecast   EmployeeForecast
	Tasks      []Task
	ChatLogs   []ChatMessage
	AIInsight  string
}

type ReassignResult struct {
	TaskID           string
	TaskTitle        string
	NewAssigneeID    string
	NewAssigneeName  string
	ProjectStatus    ProjectStatus
	Message          string
}
