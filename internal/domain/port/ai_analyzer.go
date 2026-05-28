package port

import "context"

// AIAnalyzer — GigaChat и другие LLM-провайдеры.
type AIAnalyzer interface {
	AnalyzeEmployeeContext(ctx context.Context, input EmployeeAnalysisInput) (string, error)
}

type EmployeeAnalysisInput struct {
	EmployeeName string
	Role         string
	DoneCount    int
	TotalCount   int
	DelayDays    int
	ChatLogs     []string
}
