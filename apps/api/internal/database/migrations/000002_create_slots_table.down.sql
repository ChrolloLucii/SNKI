-- Откат: удаление всех индексов и таблицы slots
DROP INDEX IF EXISTS idx_slots_filter;
DROP INDEX IF EXISTS idx_slots_starts_at;
DROP INDEX IF EXISTS idx_slots_status;
DROP TABLE IF EXISTS slots CASCADE;
