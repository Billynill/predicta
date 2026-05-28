package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/domain/port"
)

// Bot — adapter: Telegram long polling → ChatRepository.
type Bot struct {
	token      string
	chatID     string
	employees  port.EmployeeRepository
	chats      port.ChatRepository
	httpClient *http.Client
}

func NewBot(
	token, chatID string,
	employees port.EmployeeRepository,
	chats port.ChatRepository,
) *Bot {
	return &Bot{
		token:      token,
		chatID:     strings.TrimSpace(chatID),
		employees:  employees,
		chats:      chats,
		httpClient: &http.Client{Timeout: 35 * time.Second},
	}
}

func (b *Bot) Run(ctx context.Context) error {
	log.Printf("telegram bot started (chat filter: %s)", b.chatFilterLabel())

	offset := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		updates, err := b.fetchUpdates(ctx, offset)
		if err != nil {
			log.Printf("telegram getUpdates error: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		for _, update := range updates.Result {
			offset = update.UpdateID + 1
			if err := b.processUpdate(ctx, update); err != nil {
				log.Printf("telegram process update: %v", err)
			}
		}
	}
}

func (b *Bot) chatFilterLabel() string {
	if b.chatID == "" {
		return "any (set TELEGRAM_CHAT_ID after first message)"
	}
	return b.chatID
}

func (b *Bot) processUpdate(ctx context.Context, update tgUpdate) error {
	msg := update.Message
	if msg.Text == "" {
		return nil
	}

	chatIDStr := strconv.FormatInt(msg.Chat.ID, 10)
	if b.chatID != "" && chatIDStr != b.chatID {
		return nil
	}

	if b.chatID == "" {
		log.Printf("telegram: message from chat_id=%s — add TELEGRAM_CHAT_ID=%s to .env", chatIDStr, chatIDStr)
	}

	username := strings.TrimSpace(msg.From.Username)
	if username == "" {
		log.Printf("telegram: skip message without username from chat_id=%s", chatIDStr)
		return nil
	}

	sentAt := time.Unix(msg.Date, 0)
	saved, err := b.handleMessage(ctx, username, msg.Text, sentAt)
	if err != nil {
		return err
	}
	if saved {
		log.Printf("telegram: saved message from @%s", username)
	}
	return nil
}

func (b *Bot) handleMessage(ctx context.Context, username, text string, sentAt time.Time) (bool, error) {
	employee, err := b.employees.GetByTelegramNick(ctx, username)
	if err != nil {
		log.Printf("telegram: unknown user @%s (add telegram_nick to config/employees.json)", username)
		return false, nil
	}

	if err := b.chats.Save(ctx, entity.ChatMessage{
		EmployeeID: employee.ID,
		Text:       text,
		SentAt:     sentAt,
	}); err != nil {
		return false, err
	}
	return true, nil
}

type tgUpdate struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		Text string `json:"text"`
		Date int64  `json:"date"`
		From struct {
			Username string `json:"username"`
		} `json:"from"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

type updateResponse struct {
	OK     bool       `json:"ok"`
	Result []tgUpdate `json:"result"`
}

func (b *Bot) fetchUpdates(ctx context.Context, offset int) (updateResponse, error) {
	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=25",
		b.token,
		offset,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return updateResponse{}, err
	}

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return updateResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return updateResponse{}, err
	}

	var out updateResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return updateResponse{}, err
	}
	if !out.OK {
		return updateResponse{}, fmt.Errorf("telegram api error: %s", string(body))
	}
	return out, nil
}
