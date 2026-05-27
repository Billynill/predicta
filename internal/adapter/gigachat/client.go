package gigachat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/predicta/predicta/internal/domain/port"
)

const systemPrompt = `Ты — экспертный ИИ-ассистент для управления проектами и анализа рисков в команде.
Твоя задача — объяснить причину низкого темпа сотрудника на основе данных его активности.
Пиши строго, профессионально, без лишней воды. Максимум 3 предложения.`

type Client struct {
	clientID     string
	clientSecret string
	scope        string
	httpClient   *http.Client
}

func NewClient(clientID, clientSecret, scope string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		scope:        scope,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) AnalyzeEmployeeContext(ctx context.Context, input port.EmployeeAnalysisInput) (string, error) {
	userPrompt := fmt.Sprintf(
		"Сотрудник: %s (%s). Статистика Jira: Из %d задач в текущем спринте закрыта только %d. "+
			"Идет отставание от графика на %d дней. Последние логи из рабочего чата: %s. "+
			"Объясни менеджеру причину падения темпа и дай краткую рекомендацию.",
		input.EmployeeName,
		input.Role,
		input.TotalCount,
		input.DoneCount,
		input.DelayDays,
		formatLogs(input.ChatLogs),
	)

	// TODO: OAuth token + POST https://gigachat.devices.sberbank.ru/api/v1/chat/completions
	_, _ = c.buildRequest(ctx, userPrompt)
	return "", fmt.Errorf("gigachat integration not implemented; use mock AI in DEMO_MODE")
}

func (c *Client) buildRequest(ctx context.Context, userPrompt string) (*http.Request, error) {
	payload := map[string]any{
		"model": "GigaChat",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return http.NewRequestWithContext(ctx, http.MethodPost, "https://gigachat.devices.sberbank.ru/api/v1/chat/completions", bytes.NewReader(body))
}

func formatLogs(logs []string) string {
	if len(logs) == 0 {
		return "[]"
	}
	return "[" + strings.Join(logs, ", ") + "]"
}
