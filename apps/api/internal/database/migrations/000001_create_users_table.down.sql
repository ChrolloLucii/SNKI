-- Откат: удаление таблицы users и связанных индексов
DROP INDEX IF EXISTS idx_users_token;
DROP TABLE IF EXISTS users CASCADE;
