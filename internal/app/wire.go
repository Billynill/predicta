package app

import (
	"context"
	"fmt"
	"log"

	"github.com/predicta/predicta/config"
	"github.com/predicta/predicta/internal/infrastructure/gigachat"
	"github.com/predicta/predicta/internal/infrastructure/jira"
	"github.com/predicta/predicta/internal/infrastructure/memory"
	"github.com/predicta/predicta/internal/infrastructure/mock"
	"github.com/predicta/predicta/internal/infrastructure/postgres"
	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/domain/port"
	"github.com/predicta/predicta/internal/domain/service"
	"github.com/predicta/predicta/internal/usecase"
)

func buildDeps(cfg *config.Config) (*Deps, error) {
	roster, err := loadRoster(cfg)
	if err != nil {
		return nil, err
	}

	teamID := teamIDFrom(roster, cfg)
	track := "бэкенда"

	ctx := context.Background()
	members, err := roster.ListByTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}
	log.Printf("team roster: %d members from %s (team=%s)", len(members), cfg.EmployeesFile, teamID)

	velocity := service.NewVelocityEngine()
	tracker, err := buildTaskTracker(cfg, teamID, members)
	if err != nil {
		return nil, err
	}
	chats, err := buildChatRepository(cfg, members)
	if err != nil {
		return nil, err
	}
	ai := buildAI(cfg)

	project := usecase.NewProjectService(tracker, velocity, teamID, track)
	team := usecase.NewTeamService(tracker, roster, velocity, teamID)
	employee := usecase.NewEmployeeService(tracker, roster, chats, ai, velocity, teamID)
	task := usecase.NewTaskService(tracker, roster, velocity, teamID, track)

	return &Deps{
		Config:    cfg,
		TeamID:    teamID,
		Track:     track,
		Employees: roster,
		Chats:     chats,
		Tracker:   tracker,
		AI:        ai,
		Velocity:  velocity,
		Project:   project,
		Team:      team,
		Employee:  employee,
		Task:      task,
	}, nil
}

func buildTaskTracker(cfg *config.Config, teamID string, members []entity.Employee) (port.TaskTracker, error) {
	if cfg.DemoMode || cfg.TaskTrackerProvider == "mock" {
		log.Printf("task tracker: mock demo")
		return mock.NewTracker(teamID, cfg.SprintDaysRemaining, members), nil
	}

	switch cfg.TaskTrackerProvider {
	case "jira":
		log.Printf("task tracker: jira")
		return jira.NewClient(
			cfg.JiraBaseURL,
			cfg.JiraEmail,
			cfg.JiraAPIToken,
			cfg.JiraSprintID,
			cfg.JiraProjectKey,
		), nil
	default:
		return nil, fmt.Errorf("unsupported task tracker: %s", cfg.TaskTrackerProvider)
	}
}

func buildChatRepository(cfg *config.Config, members []entity.Employee) (port.ChatRepository, error) {
	if cfg.PostgresDSN != "" {
		repo, err := postgres.NewChatRepository(cfg.PostgresDSN)
		if err != nil {
			return nil, err
		}
		log.Printf("chat repository: postgres")
		return repo, nil
	}

	log.Printf("chat repository: in-memory")
	return memory.NewChatRepository(demoChatSeed(members)), nil
}

func buildAI(cfg *config.Config) port.AIAnalyzer {
	if cfg.GigaChatAuthKey != "" {
		log.Printf("ai analyzer: gigachat")
		return gigachat.NewClient(cfg.GigaChatAuthKey, cfg.GigaChatScope)
	}
	log.Printf("ai analyzer: mock fallback")
	return mock.NewAIAnalyzer()
}

func demoChatSeed(members []entity.Employee) []entity.ChatMessage {
	if len(members) < 2 {
		return nil
	}
	slow := members[1]
	return []entity.ChatMessage{
		{EmployeeID: slow.ID, Text: "Вчера 23:40: Опять сижу с этой базой данных, голова уже не варит"},
		{EmployeeID: slow.ID, Text: "Сегодня 09:15: Ребят, я дико устал, всю ночь не спал из-за семейных проблем, но постараюсь доползти до компа"},
	}
}
