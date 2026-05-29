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

	employeeSystemPrompt = `Ты — экспертный ИИ-ассистент для управления проектами и анализа рисков в команде.
На основе задач Jira и сообщений из рабочего Telegram-чата оцени загрузку сотрудника: перегружен, норма или недогруз.
Укажи, над чем человек работает сейчас. Пиши строго, профессионально, без воды. Максимум 3 предложения.`

	employeeProfileSystemPrompt = `Ты — ИИ-ассистент руководителя команды разработки.
Дай краткий вердикт по сотруднику для карточки в мобильном приложении.
Структура: 1) текущая загрузка и над чем работает; 2) риски (выгорание, срыв сроков); 3) конкретные советы менеджеру (ставить/не ставить задачи, кому передать, что уточнить в чате).
Ровно 3–4 предложения, по-русски, без воды и списков.`

	teamSystemPrompt = `Ты — эксперт по управлению agile-командой.
По данным Jira и Telegram-чатов определи загрузку каждого участника: перегружен / норма / недогруз.
Для каждого сотрудника — одно предложение: что делает и уровень загрузки.
В конце — общий вывод по балансу команды и рекомендация по перераспределению задач. Без воды, до 8 предложений.`

	projectStatusSystemPrompt = `Ты — ИИ-ассистент руководителя команды разработки.
Менеджер смотрит дашборд спринта. Дай практический совет: что сделать, чтобы успеть к дедлайну.
Если есть риск срыва — предложи 2–3 конкретных действия: перераспределить задачи, убрать scope, усилить перегруженных, ускорить review.
Если спринт идёт по плану — кратко подтверди и дай 1 профилактический совет.
3–4 предложения, по-русски, без воды и маркированных списков.`

	taskAssigneeSystemPrompt = `Ты — ИИ-ассистент руководителя команды разработки.
Менеджер хочет поставить новую задачу сотруднику. Ты должен оценить загрузку и дать чёткий совет.
Если сотрудник перегружен — категорически не рекомендуй ставить ему задачу, предупреди о риске выгорания
и предложи другого сотрудника. Обязательно укажи account_id рекомендуемого сотрудника в формате 712020:...
Пиши по-русски, 2–4 предложения, без воды.`
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
		"Сотрудник: %s (%s).\n"+
			"Спринт: закрыто %d из %d задач, отставание %d дн.\n"+
			"Задачи Jira: %s\n"+
			"Сообщения из Telegram: %s\n"+
			"Оцени загрузку и объясни, над чем человек работает и есть ли риск срыва сроков.",
		input.EmployeeName,
		input.Role,
		input.DoneCount,
		input.TotalCount,
		input.DelayDays,
		formatTasks(input.Tasks),
		formatLogs(input.ChatLogs),
	)

	return c.chat(ctx, employeeSystemPrompt, userPrompt)
}

func (c *Client) AnalyzeEmployeeProfile(ctx context.Context, input port.EmployeeProfileInput) (string, error) {
	healthLabel := mapHealthLabel(input.Health)
	userPrompt := fmt.Sprintf(
		"Сотрудник: %s (%s).\n"+
			"Статус загрузки: %s (%s).\n"+
			"Спринт: закрыто %d из %d, в работе %d.\n"+
			"Задачи Jira: %s\n"+
			"Сообщения из Telegram: %s\n"+
			"Дай вердикт и советы менеджеру.",
		input.EmployeeName,
		input.Role,
		input.Health,
		healthLabel,
		input.DoneCount,
		input.TotalCount,
		input.Remaining,
		formatTasks(input.Tasks),
		formatLogs(input.ChatLogs),
	)

	return c.chat(ctx, employeeProfileSystemPrompt, userPrompt)
}

func mapHealthLabel(health string) string {
	switch health {
	case "good":
		return "лёгкая загрузка"
	case "normal":
		return "нормальная загрузка"
	case "bad":
		return "перегруз"
	default:
		return "неизвестно"
	}
}

func (c *Client) AnalyzeTeamWorkload(ctx context.Context, input port.TeamAnalysisInput) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Спринт: %s, осталось %d дн.\n\n", input.SprintName, input.DaysLeft)
	for _, m := range input.Members {
		fmt.Fprintf(&b, "— %s (%s): закрыто %d/%d, в работе %d\n", m.Name, m.Role, m.DoneCount, m.TotalCount, m.Remaining)
		fmt.Fprintf(&b, "  Задачи: %s\n", formatTasks(m.Tasks))
		fmt.Fprintf(&b, "  Telegram: %s\n\n", formatLogs(m.ChatLogs))
	}

	return c.chat(ctx, teamSystemPrompt, b.String())
}

func (c *Client) AnalyzeProjectStatus(ctx context.Context, input port.ProjectStatusInput) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Спринт: %s (%s), осталось %d дн.\n", input.SprintName, input.TrackName, input.DaysRemaining)
	fmt.Fprintf(&b, "Прогресс: %.0f%%, закрыто %d из %d, в работе %d.\n", input.CompletionPct, input.DoneTasks, input.TotalTasks, input.RemainingTasks)
	fmt.Fprintf(&b, "Риск: %v", input.IsAtRisk)
	if input.IsAtRisk {
		fmt.Fprintf(&b, ", отставание %d дн.\n", input.DelayDays)
	} else {
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "Сообщение системы: %s\n\n", input.RiskMessage)

	fmt.Fprintf(&b, "Команда:\n")
	for _, m := range input.TeamMembers {
		fmt.Fprintf(&b, "— %s: закрыто %d/%d, в работе %d; задачи: %s\n",
			m.Name, m.DoneCount, m.TotalCount, m.Remaining, formatTasks(m.Tasks))
	}

	return c.chat(ctx, projectStatusSystemPrompt, b.String())
}

func (c *Client) RecommendTaskAssignee(ctx context.Context, input port.TaskAssigneeRecommendationInput) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Новая задача: %q\n", input.TaskTitle)
	if input.TaskDescription != "" {
		fmt.Fprintf(&b, "Описание: %s\n", input.TaskDescription)
	}
	fmt.Fprintf(&b, "Спринт: %s, осталось %d дн.\n\n", input.SprintName, input.DaysLeft)

	fmt.Fprintf(&b, "Менеджер хочет назначить: %s (account_id=%s), задач %d/%d, в работе %d\n",
		input.RequestedAssignee.Name,
		input.RequestedAssignee.AccountID,
		input.RequestedAssignee.DoneCount,
		input.RequestedAssignee.TotalCount,
		input.RequestedAssignee.Remaining,
	)
	fmt.Fprintf(&b, "Задачи: %s\n", formatTasks(input.RequestedAssignee.Tasks))

	if !input.Approved {
		fmt.Fprintf(&b, "\nРекомендуемый исполнитель: %s (account_id=%s), задач %d/%d, в работе %d\n",
			input.SuggestedAssignee.Name,
			input.SuggestedAssignee.AccountID,
			input.SuggestedAssignee.DoneCount,
			input.SuggestedAssignee.TotalCount,
			input.SuggestedAssignee.Remaining,
		)
		fmt.Fprintf(&b, "Задачи: %s\n", formatTasks(input.SuggestedAssignee.Tasks))
		fmt.Fprintf(&b, "\nСистема считает, что %s перегружен. Не назначай ему задачу.\n", input.RequestedAssignee.Name)
	} else {
		fmt.Fprintf(&b, "\nСистема считает, что загрузка %s в норме — можно назначить.\n", input.RequestedAssignee.Name)
	}

	fmt.Fprintf(&b, "\nКоманда:\n")
	for _, m := range input.TeamMembers {
		fmt.Fprintf(&b, "— %s (account_id=%s): %d/%d, в работе %d\n",
			m.Name, m.AccountID, m.DoneCount, m.TotalCount, m.Remaining)
	}

	return c.chat(ctx, taskAssigneeSystemPrompt, b.String())
}

func (c *Client) chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
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
		return "нет сообщений"
	}
	return "[" + strings.Join(logs, "; ") + "]"
}

func formatTasks(tasks []port.TaskSummary) string {
	if len(tasks) == 0 {
		return "нет задач в спринте"
	}
	parts := make([]string, 0, len(tasks))
	for _, t := range tasks {
		parts = append(parts, fmt.Sprintf("%s [%s] %s", t.ID, t.Status, t.Title))
	}
	return strings.Join(parts, "; ")
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

var _ port.AIAnalyzer = (*Client)(nil)
