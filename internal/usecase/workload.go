package usecase

import (
	"context"

	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/domain/port"
	"github.com/predicta/predicta/internal/domain/service"
)

func buildTeamMemberInputs(
	ctx context.Context,
	employees []entity.Employee,
	tasks []entity.Task,
	chats port.ChatRepository,
) ([]port.TeamMemberInput, error) {
	members := make([]port.TeamMemberInput, 0, len(employees))
	for _, emp := range employees {
		employeeTasks := filterTasksByAssignee(tasks, emp)
		done, total := countTaskProgress(employeeTasks)

		logs, err := chats.ListRecentByEmployee(ctx, emp.ID, chatLogLimit)
		if err != nil {
			return nil, err
		}

		members = append(members, port.TeamMemberInput{
			AccountID:  emp.ExternalID,
			Name:       emp.Name,
			Role:       emp.Role,
			DoneCount:  done,
			TotalCount: total,
			Remaining:  total - done,
			Tasks:      mapTasksToSummary(employeeTasks),
			ChatLogs:   formatChatLogs(logs),
		})
	}
	return members, nil
}

func teamMemberForEmployee(emp entity.Employee, members []port.TeamMemberInput) port.TeamMemberInput {
	for _, m := range members {
		if m.AccountID == emp.ExternalID || m.AccountID == emp.ID {
			return m
		}
	}
	return port.TeamMemberInput{
		AccountID:  emp.ExternalID,
		Name:       emp.Name,
		Role:       emp.Role,
		DoneCount:  0,
		TotalCount: 0,
		Remaining:  0,
	}
}

func evaluateAssigneeLoad(requested entity.Employee, velocities []entity.EmployeeVelocity) service.AssigneeLoadDecision {
	return service.EvaluateAssigneeLoad(requested, velocities)
}
