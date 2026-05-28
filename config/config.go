package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort string

	PostgresDSN string

	TaskTrackerProvider string // jira | linear | mock
	JiraBaseURL         string
	JiraEmail           string
	JiraAPIToken        string
	JiraSprintID        string
	JiraProjectKey      string

	LinearAPIKey   string
	LinearTeamID   string
	LinearSprintID string

	TelegramBotToken string
	TelegramChatID   string

	GigaChatAuthKey      string // Base64 Authorization key from Sber Studio
	GigaChatClientID     string
	GigaChatClientSecret string
	GigaChatScope        string

	SprintDaysRemaining int
	DemoMode            bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		HTTPPort:             getenv("HTTP_PORT", "8080"),
		PostgresDSN:          os.Getenv("POSTGRES_DSN"),
		TaskTrackerProvider:  getenv("TASK_TRACKER_PROVIDER", "mock"),
		JiraBaseURL:          os.Getenv("JIRA_BASE_URL"),
		JiraEmail:            os.Getenv("JIRA_EMAIL"),
		JiraAPIToken:         os.Getenv("JIRA_API_TOKEN"),
		JiraSprintID:         os.Getenv("JIRA_SPRINT_ID"),
		JiraProjectKey:       os.Getenv("JIRA_PROJECT_KEY"),
		LinearAPIKey:         os.Getenv("LINEAR_API_KEY"),
		LinearTeamID:         os.Getenv("LINEAR_TEAM_ID"),
		LinearSprintID:       os.Getenv("LINEAR_SPRINT_ID"),
		TelegramBotToken:     os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:       os.Getenv("TELEGRAM_CHAT_ID"),
		GigaChatAuthKey:      os.Getenv("GIGACHAT_AUTH_KEY"),
		GigaChatClientID:     os.Getenv("GIGACHAT_CLIENT_ID"),
		GigaChatClientSecret: os.Getenv("GIGACHAT_CLIENT_SECRET"),
		GigaChatScope:        getenv("GIGACHAT_SCOPE", "GIGACHAT_API_PERS"),
		SprintDaysRemaining:  getenvInt("SPRINT_DAYS_REMAINING", 3),
		DemoMode:             getenvBool("DEMO_MODE", true),
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getenvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func (c *Config) Addr() string {
	return fmt.Sprintf(":%s", c.HTTPPort)
}

func (c *Config) SprintEndIn() time.Duration {
	return time.Duration(c.SprintDaysRemaining) * 24 * time.Hour
}
