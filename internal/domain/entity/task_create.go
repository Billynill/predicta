package entity

type CreateTaskResult struct {
	Created               bool
	Approved              bool
	TaskID                string
	TaskTitle             string
	AssigneeID            string
	AssigneeName          string
	SuggestedAssigneeID   string
	SuggestedAssigneeName string
	AIInsight             string
}

type EmployeeProfile struct {
	Employee   Employee
	DoneCount  int
	TotalCount int
	Remaining  int
	Health     VelocityHealth
	AvatarURL  string
	Tasks      []Task
	AIInsight  string
}
