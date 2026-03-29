-- Таблица участников (связь many-to-many между users и slots)
CREATE TABLE IF NOT EXISTS participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Внешние ключи с CASCADE удалением
    slot_id UUID NOT NULL REFERENCES slots(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Статус участия (state machine: RESERVED -> PAID)
    status VARCHAR(20) NOT NULL DEFAULT 'RESERVED'
        CHECK (status IN ('RESERVED', 'PAID')),

    -- Временные метки
    reserved_at TIMESTAMP NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMP,

    -- КРИТИЧНОЕ ОГРАНИЧЕНИЕ: один пользователь не может дважды присоединиться к одному слоту
    -- Защищает от багов и race conditions на уровне БД
    CONSTRAINT unique_slot_user UNIQUE (slot_id, user_id)
);

-- Индекс для GET /me/participations (все участия пользователя)
CREATE INDEX idx_participants_user_id ON participants(user_id);

-- КРИТИЧНЫЙ: Композитный индекс для защиты от овербукинга
-- Оптимизирует запрос: SELECT COUNT(*) FROM participants
--                       WHERE slot_id = $1 AND status IN ('RESERVED','PAID')
--                       FOR UPDATE
-- При использовании FOR UPDATE PostgreSQL использует этот индекс,
-- блокируя только нужные строки (не всю таблицу)
CREATE INDEX idx_participants_slot_status ON participants(slot_id, status);

-- Индекс на FK для ускорения CASCADE операций при удалении слотов
CREATE INDEX idx_participants_slot_id ON participants(slot_id);

-- Комментарии
COMMENT ON TABLE participants IS 'Участники слотов - связь users и slots с отслеживанием статуса';
COMMENT ON COLUMN participants.status IS 'RESERVED - забронировано, PAID - оплачено и подтверждено';
COMMENT ON COLUMN participants.paid_at IS 'Время оплаты (NULL если статус RESERVED)';
COMMENT ON CONSTRAINT unique_slot_user ON participants
    IS 'Защита от повторного присоединения пользователя к одному слоту';
