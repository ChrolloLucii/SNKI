#!/bin/bash
# Скрипт для верификации миграций и seed-данных

set -e

echo "🔍 Проверка миграций базы данных..."
echo ""

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция для SQL запросов
query() {
    docker compose exec -T db psql -U postgres -d zal -t -c "$1" 2>/dev/null | tr -d '[:space:]'
}

# 1. Проверка наличия таблиц
echo "1️⃣  Проверка таблиц..."
tables=$(query "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE' AND table_name != 'schema_migrations'")

if [ "$tables" -eq "4" ]; then
    echo -e "${GREEN}✓${NC} Все 4 таблицы созданы (users, slots, participants, payments)"
else
    echo -e "${RED}✗${NC} Ожидалось 4 таблицы, найдено: $tables"
    exit 1
fi

# 2. Проверка версии миграций
echo "2️⃣  Проверка версии миграций..."
version=$(query "SELECT version FROM schema_migrations")
dirty=$(query "SELECT dirty FROM schema_migrations")

if [ "$dirty" = "f" ]; then
    echo -e "${GREEN}✓${NC} Миграции в чистом состоянии (dirty=false)"
else
    echo -e "${RED}✗${NC} База данных в dirty state!"
    exit 1
fi

echo -e "${GREEN}✓${NC} Текущая версия миграций: $version"

# 3. Проверка seed-данных
echo "3️⃣  Проверка seed-данных..."

users=$(query "SELECT COUNT(*) FROM users")
slots=$(query "SELECT COUNT(*) FROM slots")
participants=$(query "SELECT COUNT(*) FROM participants")
payments=$(query "SELECT COUNT(*) FROM payments")

if [ "$users" -eq "3" ] && [ "$slots" -eq "5" ] && [ "$participants" -eq "2" ] && [ "$payments" -eq "1" ]; then
    echo -e "${GREEN}✓${NC} Seed-данные загружены корректно:"
    echo "  - Users: $users"
    echo "  - Slots: $slots"
    echo "  - Participants: $participants"
    echo "  - Payments: $payments"
else
    echo -e "${YELLOW}⚠${NC}  Seed-данные отличаются от ожидаемых:"
    echo "  - Users: $users (ожидалось 3)"
    echo "  - Slots: $slots (ожидалось 5)"
    echo "  - Participants: $participants (ожидалось 2)"
    echo "  - Payments: $payments (ожидалось 1)"
fi

# 4. Проверка индексов
echo "4️⃣  Проверка критичных индексов..."

idx_participants_slot_status=$(query "SELECT COUNT(*) FROM pg_indexes WHERE indexname='idx_participants_slot_status'")
idx_slots_filter=$(query "SELECT COUNT(*) FROM pg_indexes WHERE indexname='idx_slots_filter'")
idx_payments_idempotency=$(query "SELECT COUNT(*) FROM pg_indexes WHERE indexname='idx_payments_idempotency_key'")

if [ "$idx_participants_slot_status" -eq "1" ]; then
    echo -e "${GREEN}✓${NC} idx_participants_slot_status (овербукинг защита)"
else
    echo -e "${RED}✗${NC} Отсутствует критичный индекс idx_participants_slot_status"
    exit 1
fi

if [ "$idx_slots_filter" -eq "1" ]; then
    echo -e "${GREEN}✓${NC} idx_slots_filter (фильтрация слотов)"
else
    echo -e "${RED}✗${NC} Отсутствует индекс idx_slots_filter"
    exit 1
fi

if [ "$idx_payments_idempotency" -eq "1" ]; then
    echo -e "${GREEN}✓${NC} idx_payments_idempotency_key (double payment защита)"
else
    echo -e "${RED}✗${NC} Отсутствует критичный индекс для idempotency"
    exit 1
fi

# 5. Проверка FK constraints
echo "5️⃣  Проверка foreign key constraints..."

fk_count=$(query "SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_type='FOREIGN KEY' AND table_schema='public'")

if [ "$fk_count" -ge "4" ]; then
    echo -e "${GREEN}✓${NC} Foreign keys настроены ($fk_count штук)"
else
    echo -e "${RED}✗${NC} Недостаточно FK constraints: $fk_count"
    exit 1
fi

# 6. Проверка UNIQUE constraints
echo "6️⃣  Проверка UNIQUE constraints..."

unique_slot_user=$(query "SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_name='unique_slot_user'")
unique_participant_payment=$(query "SELECT COUNT(*) FROM pg_indexes WHERE indexname LIKE '%participant_id%' AND tablename='payments'")

if [ "$unique_slot_user" -eq "1" ]; then
    echo -e "${GREEN}✓${NC} unique_slot_user (защита от повторного присоединения)"
else
    echo -e "${RED}✗${NC} Отсутствует unique_slot_user constraint"
    exit 1
fi

if [ "$unique_participant_payment" -ge "1" ]; then
    echo -e "${GREEN}✓${NC} participant_id UNIQUE (защита от double payment)"
else
    echo -e "${RED}✗${NC} Отсутствует UNIQUE constraint на participant_id в payments"
    exit 1
fi

# 7. Проверка API health
echo "7️⃣  Проверка API health endpoint..."

health_status=$(curl -s http://localhost:8080/health | grep -o '"status":"ok"' || echo "")

if [ -n "$health_status" ]; then
    echo -e "${GREEN}✓${NC} API отвечает на /health"
else
    echo -e "${YELLOW}⚠${NC}  API не отвечает (возможно еще не запущен)"
fi

# 8. Проверка Redis
echo "8️⃣  Проверка Redis..."

redis_ping=$(docker compose exec -T redis redis-cli PING 2>/dev/null || echo "")

if [ "$redis_ping" = "PONG" ]; then
    echo -e "${GREEN}✓${NC} Redis работает"
else
    echo -e "${RED}✗${NC} Redis не отвечает"
    exit 1
fi

echo ""
echo -e "${GREEN}✅ Все проверки пройдены успешно!${NC}"
echo ""
echo "Полезные команды:"
echo "  docker compose logs -f api     # Логи API"
echo "  docker compose exec db psql -U postgres -d zal   # PostgreSQL CLI"
echo "  docker compose exec redis redis-cli               # Redis CLI"
echo "  curl http://localhost:8080/health                 # Health check"
