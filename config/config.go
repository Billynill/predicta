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

	TaskTrackerProvider string // jira
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

	EmployeesFile string
	TeamID        string
	OpenAPIPath   string

	JWTSecret   string
	ManagerFile string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		HTTPPort:             getenv("HTTP_PORT", "8080"),
		PostgresDSN:          os.Getenv("POSTGRES_DSN"),
		TaskTrackerProvider:  getenv("TASK_TRACKER_PROVIDER", "jira"),
		JiraBaseURL:          os.Getenv("JIRA_BASE_URL"),
		JiraEmail:            os.Getenv("JIRA_EMAIL"),
		JiraAPIToken:         firstEnv("JIRA_API_TOKEN", "JIRA_API"),
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
		EmployeesFile:        getenv("EMPLOYEES_FILE", "config/employees.json"),
		TeamID:               getenv("TEAM_ID", "backend"),
		OpenAPIPath:          getenv("OPENAPI_PATH", "api/openapi.yaml"),
		JWTSecret:            getenv("JWT_SECRET", "predicta-dev-secret-change-me"),
		ManagerFile:          getenv("MANAGER_FILE", "data/manager.json"),
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return ""
}

func (c *Config) JiraEnabled() bool {
	return c.TaskTrackerProvider == "jira"
}

func (c *Config) Validate() error {
	if err := c.ValidateJira(); err != nil {
		return err
	}
	if c.GigaChatAuthKey == "" {
		return fmt.Errorf("GIGACHAT_AUTH_KEY is required")
	}
	if c.EmployeesFile == "" {
		return fmt.Errorf("EMPLOYEES_FILE is required")
	}
	return nil
}

func (c *Config) ValidateJira() error {
	if c.JiraBaseURL == "" {
		return fmt.Errorf("JIRA_BASE_URL is required")
	}
	if c.JiraEmail == "" {
		return fmt.Errorf("JIRA_EMAIL is required (email аккаунта Atlassian)")
	}
	if c.JiraAPIToken == "" {
		return fmt.Errorf("JIRA_API_TOKEN or JIRA_API is required")
	}
	if c.JiraSprintID == "" && c.JiraProjectKey == "" {
		return fmt.Errorf("set JIRA_SPRINT_ID or JIRA_PROJECT_KEY")
	}
	return nil
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

func (c *Config) Addr() string {
	return fmt.Sprintf(":%s", c.HTTPPort)
}

func (c *Config) SprintEndIn() time.Duration {
	return time.Duration(c.SprintDaysRemaining) * 24 * time.Hour
}
