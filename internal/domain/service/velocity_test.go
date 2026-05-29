package service

import (
	"testing"

	"github.com/predicta/predicta/internal/domain/entity"
)

func TestClassifyWorkloadHealth(t *testing.T) {
	tests := []struct {
		name      string
		remaining int
		metrics   TeamLoadMetrics
		want      entity.VelocityHealth
	}{
		{"zero tasks", 0, TeamLoadMetrics{5, 2.5}, entity.VelocityHealthGood},
		{"one task", 1, TeamLoadMetrics{5, 2.5}, entity.VelocityHealthGood},
		{"two tasks", 2, TeamLoadMetrics{5, 2.5}, entity.VelocityHealthNormal},
		{"three tasks typical", 3, TeamLoadMetrics{5, 3.0}, entity.VelocityHealthNormal},
		{"three tasks team leader", 3, TeamLoadMetrics{3, 2.0}, entity.VelocityHealthBad},
		{"four tasks", 4, TeamLoadMetrics{5, 3.0}, entity.VelocityHealthBad},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyWorkloadHealth(tt.remaining, tt.metrics)
			if got != tt.want {
				t.Fatalf("ClassifyWorkloadHealth(%d, %+v) = %q, want %q",
					tt.remaining, tt.metrics, got, tt.want)
			}
		})
	}
}

func TestEvaluateAssigneeLoad(t *testing.T) {
	velocities := []entity.EmployeeVelocity{
		{Employee: entity.Employee{ID: "a"}, TotalCount: 5, DoneCount: 0},
		{Employee: entity.Employee{ID: "b"}, TotalCount: 1, DoneCount: 0},
	}

	decision := EvaluateAssigneeLoad(entity.Employee{ID: "a"}, velocities)
	if decision.Approved {
		t.Fatal("expected overloaded assignee to be rejected")
	}
	if decision.Suggested.ID != "b" {
		t.Fatalf("expected suggested b, got %q", decision.Suggested.ID)
	}
}
