CREATE TABLE IF NOT EXISTS employees (
    id            TEXT PRIMARY KEY,
    external_id   TEXT NOT NULL,
    team_id       TEXT NOT NULL,
    name          TEXT NOT NULL,
    role          TEXT NOT NULL,
    telegram_nick TEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS chat_messages (
    id           BIGSERIAL PRIMARY KEY,
    employee_id  TEXT NOT NULL REFERENCES employees(id),
    message_text TEXT NOT NULL,
    sent_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_employee_sent_at
    ON chat_messages (employee_id, sent_at DESC);

-- Roster: config/employees.json (не сиды в SQL).
-- При POSTGRES_DSN чат пишется в chat_messages; employees синхронизируйте отдельно или через тот же JSON.
