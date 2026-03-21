# API Справочник

Базовый URL (локально): `http://localhost:8080`

---

## Аутентификация

В MVP используется простая демо-аутентификация через заголовки:

| Заголовок | Используется в | Значение |
|--------|---------|-------|
| `X-User-Token` | Пользовательские эндпоинты | Любой токен, соответствующий записи пользователя |
| `X-Admin-Token` | Админские эндпоинты | Значение env-переменной `ADMIN_TOKEN` |

---

## Эндпоинты

### GET `/health`

Проверка работоспособности. Аутентификация не требуется.

**Response `200`**
```json
{ "status": "ok" }
```

---

### GET `/slots`

Список доступных игровых слотов.

**Query-параметры**

| Параметр | Тип | Описание |
|-------|------|-------------|
| `sport` | string | например, `football`, `basketball` |
| `district` | string | Район города |
| `date_from` | RFC3339 | Начало диапазона дат |
| `date_to` | RFC3339 | Конец диапазона дат |

**Response `200`**
```json
[
  {
    "id": "uuid",
    "sport": "football",
    "district": "Центральный",
    "venue_name": "Стадион Динамо",
    "address": "ул. Ленина 1",
    "starts_at": "2026-04-01T18:00:00Z",
    "duration_minutes": 60,
    "capacity": 10,
    "min_players": 6,
    "deadline_at": "2026-04-01T16:00:00Z",
    "expected_price": 500,
    "max_price": 700,
    "rules_text": "Casual game, no tackles.",
    "status": "OPEN",
    "spots_left": 4
  }
]
```

---

### GET `/slots/{slotId}`

Получить детали слота.

**Response `200`** — та же структура, что и у элемента в списке выше.  
**Response `404`** — слот не найден.

---

### POST `/slots/{slotId}/join`

Забронировать место. Требуется заголовок `X-User-Token`.

Безопасно при конкуренции: используется транзакция БД с `SELECT … FOR UPDATE`.  
Возвращает `409`, если слот заполнен или пользователь уже присоединился.

**Response `201`**
```json
{
  "participant_id": "uuid",
  "slot_id": "uuid",
  "user_id": "uuid",
  "status": "RESERVED",
  "reserved_at": "2026-03-21T10:00:00Z"
}
```

**Ошибки**

| Статус | Причина |
|--------|--------|
| `401` | Отсутствует `X-User-Token` |
| `404` | Слот не найден |
| `409` | Слот заполнен или пользователь уже присоединился |

---

### POST `/slots/{slotId}/pay`

Имитация оплаты — переводит статус участника в `PAID` и создает запись платежа.  
Требуется заголовок `X-User-Token`.

**Response `200`**
```json
{
  "payment_id": "uuid",
  "participant_id": "uuid",
  "status": "PAID",
  "amount": 500,
  "currency": "RUB",
  "provider": "FAKE"
}
```

**Ошибки**

| Статус | Причина |
|--------|--------|
| `401` | Отсутствует `X-User-Token` |
| `404` | Участие не найдено |
| `409` | Уже оплачено |

---

### GET `/me/participations`

Список участий текущего пользователя. Требуется заголовок `X-User-Token`.

**Response `200`**
```json
[
  {
    "participant_id": "uuid",
    "slot": { /* same shape as GET /slots/{id} */ },
    "status": "PAID",
    "reserved_at": "2026-03-21T10:00:00Z",
    "paid_at": "2026-03-21T10:05:00Z"
  }
]
```

---

### POST `/admin/slots`

Создать новый слот. Требуется заголовок `X-Admin-Token`.

**Тело запроса**
```json
{
  "sport": "football",
  "district": "Центральный",
  "venue_name": "Стадион Динамо",
  "address": "ул. Ленина 1",
  "starts_at": "2026-04-01T18:00:00Z",
  "duration_minutes": 60,
  "capacity": 10,
  "min_players": 6,
  "deadline_at": "2026-04-01T16:00:00Z",
  "expected_price": 500,
  "max_price": 700,
  "rules_text": "Casual game, no tackles."
}
```

**Response `201`** — объект созданного слота.

**Ошибки**

| Статус | Причина |
|--------|--------|
| `401` | Отсутствует или некорректен `X-Admin-Token` |
| `422` | Ошибка валидации |

---

## Формат ошибок

Все ошибки возвращаются в едином JSON-формате:

```json
{
  "error": "human readable message"
}
```
