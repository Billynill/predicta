package app

import (
	"context"
	"fmt"
	"log"

	"github.com/predicta/predicta/config"
	"github.com/predicta/predicta/internal/delivery/http/handler"
	authinfra "github.com/predicta/predicta/internal/infrastructure/auth"
	"github.com/predicta/predicta/internal/infrastructure/gigachat"
	"github.com/predicta/predicta/internal/infrastructure/jira"
	"github.com/predicta/predicta/internal/infrastructure/memory"
	"github.com/predicta/predicta/internal/infrastructure/postgres"
	"github.com/predicta/predicta/internal/domain/port"
	"github.com/predicta/predicta/internal/domain/service"
	"github.com/predicta/predicta/internal/usecase"
)

func buildDeps(cfg *config.Config) (*Deps, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

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
	tracker, jiraSetup, err := buildTaskTracker(cfg)
	if err != nil {
		return nil, err
	}
	chats, err := buildChatRepository(cfg)
	if err != nil {
		return nil, err
	}
	ai, err := buildAI(cfg)
	if err != nil {
		return nil, err
	}

	teamSvc := usecase.NewTeamService(tracker, roster, chats, ai, velocity, teamID)
	employeeSvc := usecase.NewEmployeeService(tracker, roster, chats, userLookupFrom(tracker), ai, velocity, teamID)
	taskSvc := usecase.NewTaskService(tracker, roster, chats, ai, velocity, teamID, track)

	managers := authinfra.NewFileManagerStore(cfg.ManagerFile)
	tokens := authinfra.NewJWTService(cfg.JWTSecret)
	authSvc := usecase.NewAuthService(managers, tokens, roster, jiraSetup, teamID)

	return &Deps{
		Config:    cfg,
		TeamID:    teamID,
		Track:     track,
		Employees: roster,
		Chats:     chats,
		Tracker:   tracker,
		JiraSetup: jiraSetup,
		AI:        ai,
		Velocity:  velocity,
		Tokens:    tokens,
		Managers:  managers,
		Auth:      authSvc,
		Project:   usecase.NewProjectService(tracker, roster, chats, ai, velocity, teamID, track),
		Team:      teamSvc,
		TeamAI:    teamSvc,
		Employee:  employeeSvc,
		Profile:   employeeSvc,
		Task:      taskSvc,
		Creator:   taskSvc,
	}, nil
}

func buildTaskTracker(cfg *config.Config) (port.TaskTracker, port.JiraSetup, error) {
	switch cfg.TaskTrackerProvider {
	case "jira":
		log.Printf("task tracker: jira (sprint=%s project=%s)", cfg.JiraSprintID, cfg.JiraProjectKey)
		client := jira.NewClient(
			cfg.JiraBaseURL,
			cfg.JiraEmail,
			cfg.JiraAPIToken,
			cfg.JiraSprintID,
			cfg.JiraProjectKey,
			cfg.SprintDaysRemaining,
		)
		return client, jira.NewSetupAdapter(client), nil
	default:
		return nil, nil, fmt.Errorf("unsupported task tracker: %q (use jira)", cfg.TaskTrackerProvider)
	}
}

func buildChatRepository(cfg *config.Config) (port.ChatRepository, error) {
	if cfg.PostgresDSN != "" {
		repo, err := postgres.NewChatRepository(cfg.PostgresDSN)
		if err != nil {
			return nil, err
		}
		log.Printf("chat repository: postgres")
		return repo, nil
	}

	log.Printf("chat repository: in-memory (messages from Telegram bot)")
	return memory.NewChatRepository(nil), nil
}

func buildAI(cfg *config.Config) (port.AIAnalyzer, error) {
	log.Printf("ai analyzer: gigachat")
	return gigachat.NewClient(cfg.GigaChatAuthKey, cfg.GigaChatScope), nil
}

func userLookupFrom(tracker port.TaskTracker) port.UserLookup {
	if u, ok := tracker.(port.UserLookup); ok {
		return u
	}
	return nil
}

func buildJiraIntegration(cfg *config.Config, setup port.JiraSetup) *handler.JiraIntegration {
	if !cfg.JiraEnabled() || setup == nil {
		return nil
	}
	return handler.NewJiraIntegration(setup)
}
