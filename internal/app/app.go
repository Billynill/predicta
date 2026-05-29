package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/predicta/predicta/config"
	deliveryhttp "github.com/predicta/predicta/internal/delivery/http"
	"github.com/predicta/predicta/internal/delivery/http/handler"
	"github.com/predicta/predicta/internal/infrastructure/telegram"
)

type App struct {
	cfg       *config.Config
	server    *http.Server
	botCancel context.CancelFunc
}

func New(cfg *config.Config) (*App, error) {
	deps, err := buildDeps(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.JWTSecret == "predicta-dev-secret-change-me" {
		log.Printf("auth: using default JWT_SECRET — set JWT_SECRET in production")
	}

	apiHandler := handler.New(
		deps.Project,
		deps.Team,
		deps.TeamAI,
		deps.Employee,
		deps.Profile,
		deps.Task,
		deps.Creator,
	)
	authHandler := handler.NewAuthHandler(deps.Auth, deps.Auth, deps.Auth)
	router := deliveryhttp.NewRouter(
		apiHandler,
		authHandler,
		buildJiraIntegration(cfg, deps.JiraSetup),
		cfg.OpenAPIPath,
		deps.Tokens,
	)

	app := &App{
		cfg: cfg,
		server: &http.Server{
			Addr:              cfg.Addr(),
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}

	if cfg.TelegramBotToken != "" {
		log.Printf("telegram bot starting (chat filter: %q)", cfg.TelegramChatID)
		bot := telegram.NewBot(cfg.TelegramBotToken, cfg.TelegramChatID, deps.Employees, deps.Chats)
		ctx, cancel := context.WithCancel(context.Background())
		app.botCancel = cancel
		go func() {
			if err := bot.Run(ctx); err != nil && err != context.Canceled {
				log.Printf("telegram bot stopped: %v", err)
			}
		}()
	} else {
		log.Printf("telegram bot disabled: add TELEGRAM_BOT_TOKEN to .env")
	}

	return app, nil
}

func (a *App) Run() error {
	go func() {
		log.Printf("predicta api listening on %s (jira=%s)", a.cfg.Addr(), a.cfg.JiraProjectKey)
		if a.cfg.OpenAPIPath != "" {
			log.Printf("swagger ui: http://localhost%s/docs", a.cfg.Addr())
		}
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	if a.botCancel != nil {
		a.botCancel()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.server.Shutdown(ctx)
}
