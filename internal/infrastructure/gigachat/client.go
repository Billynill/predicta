package gigachat

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/predicta/predicta/internal/domain/port"
)

const (
	oauthURL  = "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"
	chatURL   = "https://gigachat.devices.sberbank.ru/api/v1/chat/completions"
	modelName = "GigaChat"

	systemPrompt = `Ты — экспертный ИИ-ассистент для управления проектами и анализа рисков в команде.
Твоя задача — объяснить причину низкого темпа сотрудника на основе данных его активности.
Пиши строго, профессионально, без лишней воды. Максимум 3 предложения.`
)

type Client struct {
	authKey    string
	scope      string
	httpClient *http.Client

	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

func NewClient(authKey, scope string) *Client {
	if scope == "" {
		scope = "GIGACHAT_API_PERS"
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec // required by GigaChat API certs
	}

	return &Client{
		authKey: authKey,
		scope:   scope,
		httpClient: &http.Client{
			Timeout:   60 * time.Second,
			Transport: transport,
		},
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

	return c.chat(ctx, userPrompt)
}

func (c *Client) chat(ctx context.Context, userPrompt string) (string, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return "", fmt.Errorf("gigachat auth: %w", err)
	}

	payload := chatRequest{
		Model: modelName,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.3,
		MaxTokens:   512,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gigachat chat request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		c.invalidateToken()
		return "", fmt.Errorf("gigachat chat: unauthorized")
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("gigachat chat: status %d: %s", resp.StatusCode, string(respBody))
	}

	var out chatResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("decode chat response: %w", err)
	}
	if len(out.Choices) == 0 || strings.TrimSpace(out.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("gigachat chat: empty response")
	}

	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}

func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.accessToken != "" && time.Now().Add(30*time.Second).Before(c.expiresAt) {
		token := c.accessToken
		c.mu.Unlock()
		return token, nil
	}
	c.mu.Unlock()

	return c.refreshToken(ctx)
}

func (c *Client) refreshToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Add(30*time.Second).Before(c.expiresAt) {
		return c.accessToken, nil
	}

	form := url.Values{}
	form.Set("scope", c.scope)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		oauthURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("RqUID", uuid.NewString())
	req.Header.Set("Authorization", "Basic "+c.authKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("oauth request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("oauth status %d: %s", resp.StatusCode, string(respBody))
	}

	var tokenResp oauthResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", fmt.Errorf("decode oauth response: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("oauth: empty access token")
	}

	c.accessToken = tokenResp.AccessToken
	c.expiresAt = parseExpiry(tokenResp.ExpiresAt)

	return c.accessToken, nil
}

func (c *Client) invalidateToken() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessToken = ""
	c.expiresAt = time.Time{}
}

func parseExpiry(expiresAt int64) time.Time {
	if expiresAt <= 0 {
		return time.Now().Add(29 * time.Minute)
	}
	if expiresAt > 1_000_000_000_000 {
		return time.Unix(expiresAt/1000, 0)
	}
	return time.Unix(expiresAt, 0)
}

func formatLogs(logs []string) string {
	if len(logs) == 0 {
		return "[]"
	}
	return "[" + strings.Join(logs, ", ") + "]"
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

type oauthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}
