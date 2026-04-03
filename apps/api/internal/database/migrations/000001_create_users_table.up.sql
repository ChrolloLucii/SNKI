-- Таблица пользователей с демо-токенами для аутентификации
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индекс для быстрого поиска по токену (используется в каждом запросе через X-User-Token)
CREATE INDEX idx_users_token ON users(token);

-- Комментарии для документации
COMMENT ON TABLE users IS 'Пользователи с демо-токенами для аутентификации в MVP';
COMMENT ON COLUMN users.token IS 'Демо-токен (передается в X-User-Token header)';
COMMENT ON COLUMN users.id IS 'UUID пользователя, используется как FK в других таблицах';
