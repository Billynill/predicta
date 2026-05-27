package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/domain/port"
)

// Bot слушает updates и сохраняет сообщения в ChatStore.
type Bot struct {
	token      string
	chatID     string
	employees  port.EmployeeStore
	chats      port.ChatStore
	httpClient *http.Client
}

func NewBot(token, chatID string, employees port.EmployeeStore, chats port.ChatStore) *Bot {
	return &Bot{
		token:      token,
		chatID:     chatID,
		employees:  employees,
		chats:      chats,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (b *Bot) Run(ctx context.Context) error {
	// TODO: long polling getUpdates + фильтрация по chatID и telegram nick
	<-ctx.Done()
	return ctx.Err()
}

func (b *Bot) handleMessage(ctx context.Context, username, text string, sentAt time.Time) error {
	employee, err := b.employees.GetByTelegramNick(ctx, username)
	if err != nil {
		return nil // игнорируем чужие сообщения
	}

	return b.chats.Save(ctx, entity.ChatMessage{
		EmployeeID: employee.ID,
		Text:       text,
		SentAt:     sentAt,
	})
}

type updateResponse struct {
	OK     bool `json:"ok"`
	Result []struct {
		Message struct {
			Text string `json:"text"`
			Date int64  `json:"date"`
			From struct {
				Username string `json:"username"`
			} `json:"from"`
			Chat struct {
				ID int64 `json:"id"`
			} `json:"chat"`
		} `json:"message"`
	} `json:"result"`
}

func (b *Bot) fetchUpdates(ctx context.Context, offset int) (updateResponse, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d", b.token, offset)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return updateResponse{}, err
	}
	resp, err := b.httpClient.Do(req)
	if err != nil {
		return updateResponse{}, err
	}
	defer resp.Body.Close()

	var out updateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return updateResponse{}, err
	}
	return out, nil
}
