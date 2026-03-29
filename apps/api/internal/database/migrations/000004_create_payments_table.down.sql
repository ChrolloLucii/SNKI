-- Откат: удаление всех индексов и таблицы payments
DROP INDEX IF EXISTS idx_payments_participant_id;
DROP INDEX IF EXISTS idx_payments_idempotency_key;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_provider_id;
DROP TABLE IF EXISTS payments CASCADE;
