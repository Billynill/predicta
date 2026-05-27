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

INSERT INTO employees (id, external_id, team_id, name, role, telegram_nick) VALUES
    ('emp-oleg',  'oleg-ext',  'backend', 'Олег',  'Backend', 'oleg_dev'),
    ('emp-pavel', 'pavel-ext', 'backend', 'Павел', 'Backend', 'pavel_dev')
ON CONFLICT (id) DO NOTHING;

INSERT INTO chat_messages (employee_id, message_text, sent_at) VALUES
    ('emp-pavel', 'Вчера 23:40: Опять сижу с этой базой данных, голова уже не варит', NOW() - INTERVAL '10 hours'),
    ('emp-pavel', 'Сегодня 09:15: Ребят, я дико устал, всю ночь не спал из-за семейных проблем, но постараюсь доползти до компа', NOW() - INTERVAL '2 hours')
ON CONFLICT DO NOTHING;
