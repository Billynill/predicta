package mock

import (
	"context"
	"fmt"

	"github.com/predicta/predicta/internal/domain/entity"
	"github.com/predicta/predicta/internal/domain/port"
)

const (
	TeamBackend = "backend"

	EmployeeOlegID = "emp-oleg"
	EmployeePavelID = "emp-pavel"

	SprintID = "sprint-5"
)

// Tracker — демо-данные для сценария хакатона (экраны 1–4).
type Tracker struct {
	sprintDaysRemaining int
	tasks               map[string]entity.Task
}

func NewTracker(sprintDaysRemaining int) *Tracker {
	t := &Tracker{
		sprintDaysRemaining: sprintDaysRemaining,
		tasks:               make(map[string]entity.Task),
	}
	t.seed()
	return t
}

func (t *Tracker) seed() {
	tasks := []taskSeed{
		{id: "task-1", title: "Auth API", assignee: EmployeeOlegID, status: entity.TaskStatusDone},
		{id: "task-2", title: "Users CRUD", assignee: EmployeeOlegID, status: entity.TaskStatusDone},
		{id: "task-3", title: "Notifications", assignee: EmployeeOlegID, status: entity.TaskStatusDone},
		{id: "task-4", title: "Metrics export", assignee: EmployeeOlegID, status: entity.TaskStatusDone},
		{id: "task-5", title: "Release checklist", assignee: EmployeeOlegID, status: entity.TaskStatusDone},
		{id: "task-6", title: "DB migrations", assignee: EmployeePavelID, status: entity.TaskStatusDone},
		{id: "task-7", title: "Payment gateway", assignee: EmployeePavelID, status: entity.TaskStatusInProgress},
		{id: "task-8", title: "Webhook handler", assignee: EmployeePavelID, status: entity.TaskStatusTodo},
		{id: "task-9", title: "Retry policy", assignee: EmployeePavelID, status: entity.TaskStatusTodo},
		{id: "task-10", title: "Load tests", assignee: EmployeePavelID, status: entity.TaskStatusTodo},
	}
	for _, item := range tasks {
		t.tasks[item.id] = entity.Task{
			ID:         item.id,
			ExternalID: item.id,
			Title:      item.title,
			AssigneeID: item.assignee,
			Status:     item.status,
			SprintID:   SprintID,
		}
	}
}

func (t *Tracker) GetCurrentSprint(_ context.Context, _ string) (*entity.Sprint, error) {
	return &entity.Sprint{
		ID:            SprintID,
		Name:          "Спринт №5",
		TeamID:        TeamBackend,
		DaysRemaining: t.sprintDaysRemaining,
	}, nil
}

func (t *Tracker) GetSprintTasks(_ context.Context, sprintID string) ([]entity.Task, error) {
	if sprintID != SprintID {
		return nil, fmt.Errorf("sprint not found: %s", sprintID)
	}
	out := make([]entity.Task, 0, len(t.tasks))
	for _, task := range t.tasks {
		out = append(out, task)
	}
	return out, nil
}

func (t *Tracker) ReassignTask(_ context.Context, taskExternalID, newAssigneeExternalID string) error {
	task, ok := t.tasks[taskExternalID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskExternalID)
	}
	task.AssigneeID = mapExternalAssignee(newAssigneeExternalID)
	t.tasks[taskExternalID] = task
	return nil
}

func mapExternalAssignee(externalID string) string {
	switch externalID {
	case EmployeeOlegID, "oleg-ext":
		return EmployeeOlegID
	case EmployeePavelID, "pavel-ext":
		return EmployeePavelID
	default:
		return externalID
	}
}

type taskSeed struct {
	id, title, assignee string
	status              entity.TaskStatus
}

// EmployeeStore — in-memory сотрудники для demo/mock режима.
type EmployeeStore struct{}

func NewEmployeeStore() *EmployeeStore {
	return &EmployeeStore{}
}

func (s *EmployeeStore) ListByTeam(_ context.Context, teamID string) ([]entity.Employee, error) {
	if teamID != TeamBackend {
		return nil, fmt.Errorf("team not found: %s", teamID)
	}
	return []entity.Employee{
		{ID: EmployeeOlegID, ExternalID: "oleg-ext", Name: "Олег", Role: "Backend", TelegramNick: "oleg_dev"},
		{ID: EmployeePavelID, ExternalID: "pavel-ext", Name: "Павел", Role: "Backend", TelegramNick: "pavel_dev"},
	}, nil
}

func (s *EmployeeStore) GetByID(_ context.Context, id string) (*entity.Employee, error) {
	employees, _ := s.ListByTeam(context.Background(), TeamBackend)
	for _, e := range employees {
		if e.ID == id {
			copy := e
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("employee not found: %s", id)
}

func (s *EmployeeStore) GetByTelegramNick(_ context.Context, nick string) (*entity.Employee, error) {
	employees, _ := s.ListByTeam(context.Background(), TeamBackend)
	for _, e := range employees {
		if e.TelegramNick == nick {
			copy := e
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("employee not found for nick: %s", nick)
}

// ChatStore — in-memory логи чата для demo.
type ChatStore struct {
	messages []entity.ChatMessage
}

func NewChatStore() *ChatStore {
	return &ChatStore{
		messages: seedChatMessages(),
	}
}

func seedChatMessages() []entity.ChatMessage {
	return []entity.ChatMessage{
		{EmployeeID: EmployeePavelID, Text: "Вчера 23:40: Опять сижу с этой базой данных, голова уже не варит"},
		{EmployeeID: EmployeePavelID, Text: "Сегодня 09:15: Ребят, я дико устал, всю ночь не спал из-за семейных проблем, но постараюсь доползти до компа"},
	}
}

func (s *ChatStore) Save(_ context.Context, msg entity.ChatMessage) error {
	s.messages = append(s.messages, msg)
	return nil
}

func (s *ChatStore) ListRecentByEmployee(_ context.Context, employeeID string, limit int) ([]entity.ChatMessage, error) {
	out := make([]entity.ChatMessage, 0)
	for i := len(s.messages) - 1; i >= 0 && len(out) < limit; i-- {
		if s.messages[i].EmployeeID == employeeID {
			out = append(out, s.messages[i])
		}
	}
	// reverse to chronological
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

// AIAnalyzer — заглушка с эталонным ответом из ТЗ.
type AIAnalyzer struct{}

func NewAIAnalyzer() *AIAnalyzer {
	return &AIAnalyzer{}
}

func (a *AIAnalyzer) AnalyzeEmployeeContext(_ context.Context, input port.EmployeeAnalysisInput) (string, error) {
	if input.EmployeeName == "Павел" {
		return "Падение темпа работы Павла (закрыта 1 из 5 задач) обусловлено критическим выгоранием и перегрузкой. " +
			"Анализ чата подтверждает систематические ночные переработки на фоне личных проблем. " +
			"Рекомендуется временно снизить нагрузку и перераспределить часть его задач на свободных участников команды для сохранения сроков спринта.", nil
	}
	return "Темп работы сотрудника в пределах нормы для текущего спринта. Рекомендуется сохранить текущую нагрузку.", nil
}
