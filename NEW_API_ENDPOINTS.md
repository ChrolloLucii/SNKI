# Дополнительные эндпоинты MVP (Participations & Admin Slots)

Этот документ описывает реализацию двух новых API-эндпоинтов, которые требовались для завершения контракта MVP (согласно `docs/API.md`).

## 📝 Краткое описание
Были добавлены две новые "ручки":
1. **`GET /me/participations`** — для экрана "Мои участия". Возвращает список всех слотов, в которых пользователь забронировал место или уже оплатил участие.
2. **`POST /admin/slots`** — для создания новых игровых слотов администраторами платформы.

## 📂 Где находится код

- **Контроллер "Мои участия":**
  [apps/api/internal/handlers/me.go](../apps/api/internal/handlers/me.go) — обработчик `GetMyParticipations(...)`. Выполняет SQL `JOIN` между таблицами `participants` и `slots`, чтобы одним запросом отдать и статус брони, и информацию о месте проведения.
- **Контроллер "Создание слота":**
  [apps/api/internal/handlers/admin.go](../apps/api/internal/handlers/admin.go) — обработчик `CreateSlot(...)`. Принимает JSON, валидирует обязательные текстовые поля и положительные числа (цены, вместимость), а затем делает `INSERT` в БД.
- **Роутинг:**
  [apps/api/cmd/api/main.go](../apps/api/cmd/api/main.go) — здесь прописаны маршруты `r.Get("/me/participations", ...)` и `r.Post("/admin/slots", ...)`.

## 🛡 Предварительная авторизация
Так как полноценного Middleware еще нет (запланировано в следующем этапе), контроллеры *самостоятельно* проверяют наличие необходимых токенов в заголовках:
- Для `/me/participations` требуется передать ID пользователя в заголовке `X-Demo-Token`.
- Для `/admin/slots` требуется любой непустой токен в заголовке `X-Admin-Token`.

---

## 🚀 Как тестировать

Если у вас запущен Docker (`docker compose up -d`), вы можете проверить работу API через терминал с помощью `curl`.

### 1. Создание слота от имени Администратора (POST `/admin/slots`)

```powershell
$adminJson = "{
    `"sport`": `"Basketball`",
    `"district`": `"Central`",
    `"venue_name`": `"Central Arena`",
    `"address`": `"Main St 123`",
    `"starts_at`": `"2026-05-10T18:00:00Z`",
    `"deadline_at`": `"2026-05-10T17:00:00Z`",
    `"duration_minutes`": 90,
    `"capacity`": 10,
    `"min_players`": 6,
    `"expected_price`": 500,
    `"max_price`": 700
}"

curl.exe -s -X POST `
  -H "X-Admin-Token: supersecret" `
  -H "Content-Type: application/json" `
  -d $adminJson `
  http://localhost:8080/admin/slots
```
**Ожидаемый ответ:** `201 Created` и JSON с полными данными созданного слота (с присвоенным ID и датами `created_at`/`updated_at`).

---

### 2. Получение списка своих бронирований (GET `/me/participations`)

Для проверки нужно использовать ID пользователя, который уже присоединился к какому-нибудь слоту (например, из сидов миграций: `12345678-1234-1234-1234-123456789012`).

```powershell
curl.exe -s -X GET `
  -H "X-Demo-Token: 12345678-1234-1234-1234-123456789012" `
  http://localhost:8080/me/participations
```
**Ожидаемый ответ:** `200 OK` и массив JSON-объектов, где склеены данные записи (например, `status: "PAID"`) и карточки организации (название зала, адрес, вид спорта).
