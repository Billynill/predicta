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
	ExternalID   string
	Name         string
	Role         string
	TelegramNick string
}

func AssigneeMatches(taskAssigneeID string, emp Employee) bool {
	return taskAssigneeID == emp.ID || taskAssigneeID == emp.ExternalID
}

type Task struct {
	ID          string
	ExternalID  string
	Title       string
	AssigneeID  string
	Status      TaskStatus
	StoryPoints int
	SprintID    string
}

type Sprint struct {
	ID             string
	Name           string
	TeamID         string
	DaysRemaining  int
	TotalTasks     int
	CompletedTasks int
}

type ChatMessage struct {
	ID         string
	EmployeeID string
	Text       string
	SentAt     time.Time
}
