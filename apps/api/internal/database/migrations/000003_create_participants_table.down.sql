-- Откат: удаление всех индексов и таблицы participants
DROP INDEX IF EXISTS idx_participants_user_id;
DROP INDEX IF EXISTS idx_participants_slot_status;
DROP INDEX IF EXISTS idx_participants_slot_id;
DROP TABLE IF EXISTS participants CASCADE;
