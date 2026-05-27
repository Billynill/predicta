package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/predicta/predicta/config"
	httpadapter "github.com/predicta/predicta/internal/adapter/http"
	"github.com/predicta/predicta/internal/adapter/http/handler"
	"github.com/predicta/predicta/internal/adapter/gigachat"
	"github.com/predicta/predicta/internal/adapter/jira"
	"github.com/predicta/predicta/internal/adapter/mock"
	"github.com/predicta/predicta/internal/domain/port"
	"github.com/predicta/predicta/internal/domain/service"
	"github.com/predicta/predicta/internal/usecase"
)

type App struct {
	cfg    *config.Config
	server *http.Server
}

func New(cfg *config.Config) (*App, error) {
	tracker, employees, chats, ai, err := buildAdapters(cfg)
	if err != nil {
		return nil, err
	}

	velocity := service.NewVelocityEngine()
	projectSvc := usecase.NewProjectService(tracker, velocity, mock.TeamBackend, "бэкенда")
	teamSvc := usecase.NewTeamService(tracker, employees, velocity, mock.TeamBackend)
	employeeSvc := usecase.NewEmployeeService(tracker, employees, chats, ai, velocity, mock.TeamBackend)
	taskSvc := usecase.NewTaskService(tracker, employees, velocity, projectSvc, mock.TeamBackend, "бэкенда")

	h := handler.New(projectSvc, teamSvc, employeeSvc, taskSvc)
	router := httpadapter.NewRouter(h)

	return &App{
		cfg: cfg,
		server: &http.Server{
			Addr:              cfg.Addr(),
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}, nil
}

func buildAdapters(cfg *config.Config) (port.TaskTracker, port.EmployeeStore, port.ChatStore, port.AIAnalyzer, error) {
	if cfg.DemoMode || cfg.TaskTrackerProvider == "mock" {
		return mock.NewTracker(cfg.SprintDaysRemaining),
			mock.NewEmployeeStore(),
			mock.NewChatStore(),
			mock.NewAIAnalyzer(),
			nil
	}

	var tracker port.TaskTracker
	switch cfg.TaskTrackerProvider {
	case "jira":
		tracker = jira.NewClient(cfg.JiraBaseURL, cfg.JiraEmail, cfg.JiraAPIToken, cfg.JiraSprintID, cfg.JiraProjectKey)
	default:
		return nil, nil, nil, nil, fmt.Errorf("unsupported task tracker: %s", cfg.TaskTrackerProvider)
	}

	// TODO: wire postgres store when POSTGRES_DSN is set
	employees := mock.NewEmployeeStore()
	chats := mock.NewChatStore()
	ai := gigachat.NewClient(cfg.GigaChatClientID, cfg.GigaChatClientSecret, cfg.GigaChatScope)

	return tracker, employees, chats, ai, nil
}

func (a *App) Run() error {
	go func() {
		log.Printf("predicta api listening on %s (demo=%v)", a.cfg.Addr(), a.cfg.DemoMode)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.server.Shutdown(ctx)
}
