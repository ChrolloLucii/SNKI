-- Таблица платежей с защитой от двойной оплаты
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- КРИТИЧНОЕ: один платеж на участника (защита от double payment на уровне БД)
    participant_id UUID NOT NULL UNIQUE
        REFERENCES participants(id) ON DELETE CASCADE,

    -- КРИТИЧНОЕ: idempotency key для защиты от повторных запросов
    -- Клиент генерирует UUID и передает в X-Idempotency-Key header
    -- При повторном запросе с тем же ключом возвращаем cached response
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,

    -- Информация о платеже
    amount INTEGER NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    provider VARCHAR(50) NOT NULL DEFAULT 'FAKE',
    status VARCHAR(20) NOT NULL DEFAULT 'PAID'
        CHECK (status IN ('PENDING', 'PAID', 'FAILED', 'REFUNDED')),

    -- Метаданные провайдера (для production интеграций)
    provider_payment_id VARCHAR(255),  -- ID платежа в системе провайдера
    provider_metadata JSONB,           -- Дополнительная информация от провайдера

    -- Временные метки
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индекс для поиска платежа по участнику (POST /slots/{id}/pay)
CREATE INDEX idx_payments_participant_id ON payments(participant_id);

-- Индекс для поиска по idempotency key (проверка дублей)
CREATE INDEX idx_payments_idempotency_key ON payments(idempotency_key);

-- Индекс для отчетов по статусу платежей
CREATE INDEX idx_payments_status ON payments(status);

-- Индекс для поиска по ID провайдера (для webhook обработки)
CREATE INDEX idx_payments_provider_id ON payments(provider_payment_id)
    WHERE provider_payment_id IS NOT NULL;

-- Комментарии
COMMENT ON TABLE payments IS 'Платежи за участие в слотах (MVP использует FAKE провайдер)';
COMMENT ON COLUMN payments.participant_id IS 'UNIQUE FK - один платеж на участника (защита от double payment)';
COMMENT ON COLUMN payments.idempotency_key IS 'Ключ идемпотентности для защиты от повторных запросов (генерируется на клиенте)';
COMMENT ON COLUMN payments.provider IS 'FAKE для MVP, в production: STRIPE, YOOKASSA, и т.д.';
COMMENT ON COLUMN payments.provider_payment_id IS 'ID платежа в системе провайдера (для сверки и webhooks)';
COMMENT ON COLUMN payments.status IS 'PENDING - в процессе, PAID - успешно, FAILED - ошибка, REFUNDED - возврат';
