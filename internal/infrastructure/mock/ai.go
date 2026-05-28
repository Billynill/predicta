package mock

import (
	"context"

	"github.com/predicta/predicta/internal/domain/port"
)

// AIAnalyzer — заглушка, если GigaChat не настроен.
type AIAnalyzer struct{}

func NewAIAnalyzer() *AIAnalyzer {
	return &AIAnalyzer{}
}

func (a *AIAnalyzer) AnalyzeEmployeeContext(_ context.Context, input port.EmployeeAnalysisInput) (string, error) {
	if input.TotalCount > 0 && input.DoneCount < input.TotalCount/2 {
		return "Падение темпа сотрудника обусловлено перегрузкой и признаками выгорания по данным чата. " +
			"Рекомендуется временно снизить нагрузку и перераспределить часть задач на свободных участников команды.", nil
	}
	return "Темп работы сотрудника в пределах нормы для текущего спринта. Рекомендуется сохранить текущую нагрузку.", nil
}
