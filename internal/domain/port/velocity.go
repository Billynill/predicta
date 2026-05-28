package port

import "github.com/predicta/predicta/internal/domain/entity"

// VelocityCalculator — доменная математика темпа (без I/O).
type VelocityCalculator interface {
	BuildProjectStatus(sprint entity.Sprint, tasks []entity.Task, trackName string) entity.ProjectStatus
	BuildTeamVelocity(employees []entity.Employee, tasks []entity.Task) []entity.EmployeeVelocity
	BuildEmployeeForecast(employee entity.Employee, tasks []entity.Task, sprintDaysLeft int) entity.EmployeeForecast
}
