-- Таблица игровых слотов (временные окна для игр)
CREATE TABLE IF NOT EXISTS slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Основная информация о месте и виде спорта
    sport VARCHAR(100) NOT NULL,
    district VARCHAR(100) NOT NULL,
    venue_name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,

    -- Временные параметры
    starts_at TIMESTAMP NOT NULL,
    deadline_at TIMESTAMP NOT NULL,
    duration_minutes INTEGER NOT NULL CHECK (duration_minutes > 0),

    -- Параметры вместимости и ценообразования
    capacity INTEGER NOT NULL CHECK (capacity > 0),
    min_players INTEGER NOT NULL CHECK (min_players > 0 AND min_players <= capacity),
    expected_price INTEGER NOT NULL CHECK (expected_price >= 0),
    max_price INTEGER NOT NULL CHECK (max_price >= expected_price),

    -- Правила и статус слота
    rules_text TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN'
        CHECK (status IN ('OPEN', 'CANCELLED', 'COMPLETED')),

    -- Метаданные
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- КРИТИЧНЫЙ: Композитный индекс для фильтрации GET /slots
-- Покрывает запросы: ?sport=football&district=Центральный&date_from=...
-- Partial index для оптимизации (индексируем только открытые слоты)
CREATE INDEX idx_slots_filter ON slots(sport, district, starts_at)
    WHERE status = 'OPEN';

-- Индекс для сортировки по времени начала
CREATE INDEX idx_slots_starts_at ON slots(starts_at);

-- Индекс для админских запросов по статусу
CREATE INDEX idx_slots_status ON slots(status);

-- Комментарии
COMMENT ON TABLE slots IS 'Игровые слоты - временные окна для организации игр';
COMMENT ON COLUMN slots.deadline_at IS 'Крайний срок регистрации (должен быть раньше starts_at)';
COMMENT ON COLUMN slots.expected_price IS 'Цена за участие при полном заполнении слота';
COMMENT ON COLUMN slots.max_price IS 'Максимальная цена при минимальном количестве игроков';
COMMENT ON COLUMN slots.status IS 'OPEN - открыт для регистрации, CANCELLED - отменен, COMPLETED - завершен';
