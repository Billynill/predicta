package usecase

import (
	"context"
	"fmt"

	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/domain/port"
)

const (
	defaultTeamID = "backend"
	defaultTrack  = "бэкенда"
	chatLogLimit  = 10
)

type sprintContext struct {
	sprint *entity.Sprint
	tasks  []entity.Task
}

func loadSprintContext(
	ctx context.Context,
	tracker port.TaskTracker,
	teamID string,
) (*sprintContext, error) {
	sprint, err := tracker.GetCurrentSprint(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("get sprint: %w", err)
	}

	tasks, err := tracker.GetSprintTasks(ctx, sprint.ID)
	if err != nil {
		return nil, fmt.Errorf("get sprint tasks: %w", err)
	}

	return &sprintContext{sprint: sprint, tasks: tasks}, nil
}

func resolveTeamID(teamID string) string {
	if teamID == "" {
		return defaultTeamID
	}
	return teamID
}

func resolveTrack(track string) string {
	if track == "" {
		return defaultTrack
	}
	return track
}
