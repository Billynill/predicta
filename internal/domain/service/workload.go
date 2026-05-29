package service

import "github.com/predicta/predicta/internal/domain/entity"

// TeamLoadMetrics — агрегаты загрузки команды в спринте.
type TeamLoadMetrics struct {
	MaxRemaining int
	AvgRemaining float64
}

// ComputeTeamLoadMetrics — max/avg открытых задач по velocity-срезу.
func ComputeTeamLoadMetrics(velocities []entity.EmployeeVelocity) TeamLoadMetrics {
	totalRemaining := 0
	maxRemaining := 0

	for _, v := range velocities {
		remaining := v.TotalCount - v.DoneCount
		totalRemaining += remaining
		if remaining > maxRemaining {
			maxRemaining = remaining
		}
	}

	avg := 0.0
	if len(velocities) > 0 {
		avg = float64(totalRemaining) / float64(len(velocities))
	}

	return TeamLoadMetrics{
		MaxRemaining: maxRemaining,
		AvgRemaining: avg,
	}
}

// ClassifyWorkloadHealth — good / normal / bad для UI.
func ClassifyWorkloadHealth(remaining int, metrics TeamLoadMetrics) entity.VelocityHealth {
	if remaining <= 1 {
		return entity.VelocityHealthGood
	}
	if remaining >= 4 {
		return entity.VelocityHealthBad
	}
	if remaining >= 3 && remaining == metrics.MaxRemaining && float64(remaining) > metrics.AvgRemaining {
		return entity.VelocityHealthBad
	}
	return entity.VelocityHealthNormal
}

// AssigneeLoadDecision — результат проверки перед назначением новой задачи.
type AssigneeLoadDecision struct {
	Approved  bool
	Suggested entity.Employee
}

// EvaluateAssigneeLoad — можно ли назначить задачу исполнителю; если нет — кого предложить.
func EvaluateAssigneeLoad(requested entity.Employee, velocities []entity.EmployeeVelocity) AssigneeLoadDecision {
	metrics := ComputeTeamLoadMetrics(velocities)

	var requestedLoad *entity.EmployeeVelocity
	minRemaining := -1
	var (
		lightest    entity.Employee
		hasLightest bool
	)

	for i := range velocities {
		v := &velocities[i]
		remaining := v.TotalCount - v.DoneCount

		if v.Employee.ID == requested.ID {
			vCopy := *v
			requestedLoad = &vCopy
		}

		if !hasLightest || remaining < minRemaining {
			minRemaining = remaining
			lightest = v.Employee
			hasLightest = true
		}
	}

	if requestedLoad == nil {
		return AssigneeLoadDecision{Approved: true}
	}

	requestedRemaining := requestedLoad.TotalCount - requestedLoad.DoneCount
	if !isOverloadedForAssignment(requestedRemaining, metrics) {
		return AssigneeLoadDecision{Approved: true}
	}

	suggested := lightest
	if suggested.ID == requested.ID {
		for i := range velocities {
			candidate := velocities[i].Employee
			if candidate.ID == requested.ID {
				continue
			}
			suggested = candidate
			break
		}
	}

	return AssigneeLoadDecision{
		Approved:  false,
		Suggested: suggested,
	}
}

func isOverloadedForAssignment(remaining int, metrics TeamLoadMetrics) bool {
	if remaining >= 3 {
		return true
	}
	return remaining >= 2 &&
		remaining == metrics.MaxRemaining &&
		float64(remaining) > metrics.AvgRemaining
}
