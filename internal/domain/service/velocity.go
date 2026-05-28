package service

import (
	"math"

	"github.com/predicta/predicta/internal/domain/entity"
)

const defaultVelocityPerDay = 1.0

// VelocityEngine — чистая доменная логика расчёта темпа (без I/O).
type VelocityEngine struct{}

func NewVelocityEngine() *VelocityEngine {
	return &VelocityEngine{}
}

func (e *VelocityEngine) BuildProjectStatus(
	sprint entity.Sprint,
	tasks []entity.Task,
	trackName string,
) entity.ProjectStatus {
	remaining := countRemaining(tasks)
	avgVelocity := e.teamAvgVelocityPerDay(tasks, sprint.DaysRemaining)

	delayDays := 0
	isAtRisk := false
	if avgVelocity > 0 && float64(remaining)/avgVelocity > float64(sprint.DaysRemaining) {
		neededDays := int(math.Ceil(float64(remaining) / avgVelocity))
		delayDays = neededDays - sprint.DaysRemaining
		isAtRisk = delayDays > 0
	}

	total := len(tasks)
	done := countDone(tasks)
	pct := 0.0
	if total > 0 {
		pct = float64(done) / float64(total) * 100
	}

	msg := "Спринт идёт по плану"
	if isAtRisk {
		msg = sprint.Name + ". Риск срыва дедлайна " + trackName + " на " + itoa(delayDays) + " дн."
	}

	return entity.ProjectStatus{
		SprintName:    sprint.Name,
		CompletionPct: pct,
		DelayDays:     delayDays,
		IsAtRisk:      isAtRisk,
		RiskMessage:   msg,
		TrackName:     trackName,
		DaysRemaining: sprint.DaysRemaining,
	}
}

func (e *VelocityEngine) BuildTeamVelocity(
	employees []entity.Employee,
	tasks []entity.Task,
) []entity.EmployeeVelocity {
	result := make([]entity.EmployeeVelocity, 0, len(employees))

	for _, emp := range employees {
		done, total := countByAssignee(tasks, emp)
		health := entity.VelocityHealthGood
		if total > 0 && done < total/2 {
			health = entity.VelocityHealthBad
		}

		result = append(result, entity.EmployeeVelocity{
			Employee:   emp,
			DoneCount:  done,
			TotalCount: total,
			Health:     health,
		})
	}

	return result
}

func (e *VelocityEngine) BuildEmployeeForecast(
	employee entity.Employee,
	tasks []entity.Task,
	sprintDaysLeft int,
) entity.EmployeeForecast {
	remaining := 0
	for _, t := range tasks {
		if entity.AssigneeMatches(t.AssigneeID, employee) && t.Status != entity.TaskStatusDone {
			remaining++
		}
	}

	velocity := e.employeeVelocityPerDay(tasks, employee)
	daysToComplete := sprintDaysLeft
	if velocity > 0 {
		daysToComplete = int(math.Ceil(float64(remaining) / velocity))
	}
	if remaining == 0 {
		daysToComplete = 0
	}

	delay := daysToComplete - sprintDaysLeft
	if delay < 0 {
		delay = 0
	}

	return entity.EmployeeForecast{
		EmployeeID:     employee.ID,
		EmployeeName:   employee.Name,
		RemainingTasks: remaining,
		DaysToComplete: daysToComplete,
		SprintDaysLeft: sprintDaysLeft,
		DelayDays:      delay,
	}
}

func (e *VelocityEngine) teamAvgVelocityPerDay(tasks []entity.Task, sprintDaysLeft int) float64 {
	if sprintDaysLeft <= 0 {
		sprintDaysLeft = 1
	}
	done := countDone(tasks)
	return float64(done) / float64(sprintDaysLeft)
}

func (e *VelocityEngine) employeeVelocityPerDay(tasks []entity.Task, employee entity.Employee) float64 {
	done, _ := countByAssignee(tasks, employee)
	if done == 0 {
		return defaultVelocityPerDay * 0.2 // консервативный прогноз при нулевом темпе
	}
	return float64(done) / defaultVelocityPerDay
}

func countRemaining(tasks []entity.Task) int {
	n := 0
	for _, t := range tasks {
		if t.Status != entity.TaskStatusDone {
			n++
		}
	}
	return n
}

func countDone(tasks []entity.Task) int {
	n := 0
	for _, t := range tasks {
		if t.Status == entity.TaskStatusDone {
			n++
		}
	}
	return n
}

func countByAssignee(tasks []entity.Task, emp entity.Employee) (done, total int) {
	for _, t := range tasks {
		if !entity.AssigneeMatches(t.AssigneeID, emp) {
			continue
		}
		total++
		if t.Status == entity.TaskStatusDone {
			done++
		}
	}
	return done, total
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
