package port

import "context"

type TaskAssigneeRecommendationInput struct {
	TaskTitle           string
	TaskDescription     string
	SprintName          string
	DaysLeft            int
	Approved            bool
	RequestedAssignee   TeamMemberInput
	SuggestedAssignee   TeamMemberInput
	TeamMembers         []TeamMemberInput
}

// AIAnalyzer — GigaChat и другие LLM-провайдеры.
type AIAnalyzer interface {
	AnalyzeEmployeeContext(ctx context.Context, input EmployeeAnalysisInput) (string, error)
	AnalyzeEmployeeProfile(ctx context.Context, input EmployeeProfileInput) (string, error)
	AnalyzeTeamWorkload(ctx context.Context, input TeamAnalysisInput) (string, error)
	AnalyzeProjectStatus(ctx context.Context, input ProjectStatusInput) (string, error)
	RecommendTaskAssignee(ctx context.Context, input TaskAssigneeRecommendationInput) (string, error)
}

type TaskSummary struct {
	ID     string
	Title  string
	Status string
}

type EmployeeAnalysisInput struct {
	EmployeeName string
	Role         string
	DoneCount    int
	TotalCount   int
	DelayDays    int
	Tasks        []TaskSummary
	ChatLogs     []string
}

type EmployeeProfileInput struct {
	EmployeeName string
	Role         string
	Health       string
	DoneCount    int
	TotalCount   int
	Remaining    int
	Tasks        []TaskSummary
	ChatLogs     []string
}

type TeamMemberInput struct {
	AccountID  string
	Name       string
	Role       string
	DoneCount  int
	TotalCount int
	Remaining  int
	Tasks      []TaskSummary
	ChatLogs   []string
}

type TeamAnalysisInput struct {
	SprintName string
	DaysLeft   int
	Members    []TeamMemberInput
}

type ProjectStatusInput struct {
	SprintName      string
	TrackName       string
	DaysRemaining   int
	CompletionPct   float64
	DelayDays       int
	IsAtRisk        bool
	RiskMessage     string
	RemainingTasks  int
	TotalTasks      int
	DoneTasks       int
	TeamMembers     []TeamMemberInput
}
