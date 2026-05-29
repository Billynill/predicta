# Predicta API — Android

## Подключение

| | |
|--|--|
| Base URL | `https://<ngrok-id>.ngrok-free.app/` — URL выдаёт бекенд-команда |
| Content-Type | `application/json` (POST) |

**Заголовки на каждый запрос:**

```
ngrok-skip-browser-warning: true
```

**Заголовок на все `/api/*`, кроме register и login:**

```
Authorization: Bearer <token>
```

---

## Enum

### `health` — загрузка сотрудника

| Значение | Смысл |
|----------|--------|
| `good` | 0–1 открытая задача |
| `normal` | 2–3 задачи |
| `bad` | 4+ задач |

### `status` — статус задачи

| Значение | Смысл |
|----------|--------|
| `todo` | Не начата |
| `in_progress` | В работе |
| `done` | Закрыта |

### Ошибка

```json
{ "error": "текст ошибки" }
```

| Код | Когда |
|-----|--------|
| `400` | Невалидное тело |
| `401` | Нет/неверный JWT |
| `409` | Начальник уже зарегистрирован |
| `500` | Ошибка сервера |

---

## Ручки

### `GET /health`

Без авторизации.

**Response `200`:**
```json
{ "status": "ok" }
```

---

### `POST /api/auth/register`

Без авторизации. Один раз.

**Request:**
```json
{
  "first_name": "Иван",
  "last_name": "Иванов",
  "email": "boss@company.com",
  "password": "secret123",
  "telegram_nick": "@dic0ntr0L",
  "phone": "+79001234567",
  "avatar_url": "https://example.com/avatar.jpg"
}
```

| Поле | Обяз. | Тип |
|------|-------|-----|
| `first_name` | да | string |
| `last_name` | да | string |
| `email` | да | string |
| `password` | да | string, min 6 |
| `telegram_nick` | да | string |
| `phone` | да | string |
| `avatar_url` | нет | string |

**Response `201`:**
```json
{ "message": "manager registered" }
```

---

### `POST /api/auth/login`

Без авторизации.

**Request:**
```json
{
  "email": "boss@company.com",
  "password": "secret123"
}
```

| Поле | Обяз. | Тип |
|------|-------|-----|
| `email` | да | string |
| `password` | да | string |

**Response `200`:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

---

### `GET /api/auth/me`

**Response `200`:**
```json
{
  "first_name": "Иван",
  "last_name": "Иванов",
  "email": "boss@company.com",
  "telegram_nick": "dic0ntr0L",
  "phone": "+79001234567",
  "subordinates_count": 4,
  "avatar_url": "https://example.com/avatar.jpg",
  "jira_display_name": "Alexander",
  "jira_email": "eeboy159@gmail.com"
}
```

| Поле | Тип |
|------|-----|
| `first_name` | string |
| `last_name` | string |
| `email` | string |
| `telegram_nick` | string |
| `phone` | string |
| `subordinates_count` | int |
| `avatar_url` | string |
| `jira_display_name` | string |
| `jira_email` | string |

---

### `GET /api/project/status`

**Response `200`:**
```json
{
  "sprint_name": "HP Спринт 1",
  "completion_pct": 9.09,
  "delay_days": 33,
  "is_at_risk": true,
  "risk_message": "HP Спринт 1. Риск срыва дедлайна бэкенда на 33 дн.",
  "ai_advice": "Перераспределите задачи...",
  "track_name": "бэкенда",
  "days_remaining": 14
}
```

| Поле | Тип |
|------|-----|
| `sprint_name` | string |
| `completion_pct` | number |
| `delay_days` | int |
| `is_at_risk` | bool |
| `risk_message` | string |
| `ai_advice` | string |
| `track_name` | string |
| `days_remaining` | int |

---

### `GET /api/team/velocity`

**Response `200`:** массив
```json
[
  {
    "id": "712020:ec2b01dc-3023-46c8-b8b0-b55fa31c1a8c",
    "name": "Attacking",
    "role": "Developer",
    "done_count": 0,
    "total_count": 5,
    "health": "bad"
  }
]
```

| Поле | Тип |
|------|-----|
| `id` | string (Jira account_id) |
| `name` | string |
| `role` | string |
| `done_count` | int |
| `total_count` | int |
| `health` | `good` \| `normal` \| `bad` |

---

### `GET /api/team/insights`

**Response `200`:**
```json
{
  "ai_insight": "Attacking перегружен..."
}
```

| Поле | Тип |
|------|-----|
| `ai_insight` | string |

---

### `GET /api/employee/{id}`

`{id}` — Jira account_id из `team/velocity`.

**Response `200`:**
```json
{
  "id": "712020:bf8d95b2-ee76-4144-882d-5c0d80ff697a",
  "name": "Medved",
  "role": "Developer",
  "telegram_nick": "RustamAvezov",
  "avatar_url": "https://secure.gravatar.com/avatar/...",
  "done_count": 0,
  "total_count": 1,
  "remaining_count": 1,
  "health": "good",
  "ai_insight": "Medved недогружен...",
  "tasks": [
    {
      "id": "HP-20",
      "title": "Интеграция GigaChat",
      "status": "in_progress"
    }
  ]
}
```

| Поле | Тип |
|------|-----|
| `id` | string |
| `name` | string |
| `role` | string |
| `telegram_nick` | string |
| `avatar_url` | string |
| `done_count` | int |
| `total_count` | int |
| `remaining_count` | int |
| `health` | `good` \| `normal` \| `bad` |
| `ai_insight` | string |
| `tasks` | Task[] |

**Task:**

| Поле | Тип |
|------|-----|
| `id` | string |
| `title` | string |
| `status` | `todo` \| `in_progress` \| `done` |

---

### `GET /api/employee/{id}/analytics`

**Response `200`:**
```json
{
  "employee_id": "712020:ec2b01dc-3023-46c8-b8b0-b55fa31c1a8c",
  "employee_name": "Attacking",
  "role": "Developer",
  "forecast_days_to_complete": 25,
  "sprint_days_left": 14,
  "delay_days": 11,
  "ai_insight": "Низкий темп...",
  "tasks": [
    { "id": "HP-12", "title": "...", "status": "todo" }
  ]
}
```

| Поле | Тип |
|------|-----|
| `employee_id` | string |
| `employee_name` | string |
| `role` | string |
| `forecast_days_to_complete` | int |
| `sprint_days_left` | int |
| `delay_days` | int |
| `ai_insight` | string |
| `tasks` | Task[] |

---

### `POST /api/tasks/create`

**Request:**
```json
{
  "title": "Новая фича API",
  "description": "Добавить endpoint для отчётов",
  "assignee_id": "712020:ec2b01dc-3023-46c8-b8b0-b55fa31c1a8c",
  "force": false
}
```

| Поле | Обяз. | Тип |
|------|-------|-----|
| `title` | да | string |
| `description` | нет | string |
| `assignee_id` | да | string |
| `force` | нет | bool |

**Response `200` — задача не создана (перегруз):**
```json
{
  "created": false,
  "approved": false,
  "assignee_id": "712020:ec2b01dc-3023-46c8-b8b0-b55fa31c1a8c",
  "assignee_name": "Attacking",
  "suggested_assignee_id": "712020:f82e4007-6a4d-447c-8567-153896529978",
  "suggested_assignee_name": "Taras",
  "ai_insight": "Attacking перегружен..."
}
```

**Response `200` — задача создана:**
```json
{
  "created": true,
  "approved": true,
  "task_id": "HP-23",
  "task_title": "Новая фича API",
  "assignee_id": "712020:f82e4007-6a4d-447c-8567-153896529978",
  "assignee_name": "Taras",
  "ai_insight": "Загрузка в норме, задача создана."
}
```

| Поле | Тип |
|------|-----|
| `created` | bool |
| `approved` | bool |
| `task_id` | string? |
| `task_title` | string? |
| `assignee_id` | string |
| `assignee_name` | string |
| `suggested_assignee_id` | string? |
| `suggested_assignee_name` | string? |
| `ai_insight` | string |

---

### `POST /api/tasks/reassign`

**Request:**
```json
{
  "task_id": "HP-12",
  "new_executor_id": "712020:bf8d95b2-ee76-4144-882d-5c0d80ff697a"
}
```

| Поле | Обяз. | Тип |
|------|-------|-----|
| `task_id` | да | string |
| `new_executor_id` | да | string |

**Response `200`:**
```json
{
  "message": "Задача перенаправлена Medved...",
  "project_status": {
    "sprint_name": "HP Спринт 1",
    "completion_pct": 9.09,
    "delay_days": 28,
    "is_at_risk": true,
    "risk_message": "...",
    "ai_advice": "",
    "track_name": "бэкенда",
    "days_remaining": 14
  }
}
```

| Поле | Тип |
|------|-----|
| `message` | string |
| `project_status` | ProjectStatus (см. `GET /api/project/status`) |
