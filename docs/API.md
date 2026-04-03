# API контракт для фронтенда (MVP)

Этот документ нужен, чтобы фронтенд уже мог работать по понятному контракту, даже пока часть бэкенда еще не дописана.

Важно: сейчас в коде реально поднят только GET /health. Остальные ручки в этом документе являются целевым контрактом, который нужно реализовать в бэкенде.

## Базовые правила

- Базовый URL локально: http://localhost:8080
- Формат обмена: JSON
- Время и даты: ISO-8601 в UTC
- Идентификаторы: UUID
- Деньги: целое число в рублях (например, 500 = 500 рублей)

## Заголовки

- Для публичных ручек токен не нужен.
- Для пользовательских ручек нужен заголовок X-Demo-Token.
- Для админских ручек нужен заголовок X-Admin-Token.
- Для оплаты нужен заголовок X-Idempotency-Key (уникальный ключ запроса оплаты, чтобы случайно не оплатить дважды).

## Что фронту делать по шагам

1. На экране списка игр дергать GET /slots.
2. При открытии карточки игры дергать GET /slots/{slotId}.
3. При нажатии "Участвовать" дергать POST /slots/{slotId}/join.
4. При нажатии "Оплатить" дергать POST /slots/{slotId}/pay.
5. В личном кабинете дергать GET /me/participations.

## Единый формат ошибок

Если что-то пошло не так, API должен возвращать:

- status: HTTP-статус ошибки
- code: короткий машинный код ошибки
- message: понятный текст для человека

Примеры code:

- UNAUTHORIZED
- SLOT_NOT_FOUND
- SLOT_FULL
- ALREADY_JOINED
- ALREADY_PAID
- VALIDATION_ERROR
- INTERNAL_ERROR

## Модель данных: слот

Поля, которые фронт должен получать в слоте:

- id
- sport
- district
- venue_name
- address
- starts_at
- deadline_at
- duration_minutes
- capacity
- min_players
- expected_price
- max_price
- rules_text
- status (OPEN, CANCELLED, COMPLETED)
- created_at
- updated_at

Дополнительно для удобства UI желательно отдавать:

- current_participants
- free_spots

## Модель данных: участие

- id
- slot_id
- user_id
- status (RESERVED, PAID)
- reserved_at
- paid_at

## Модель данных: платеж

- id
- participant_id
- idempotency_key
- amount
- currency
- provider
- status (PENDING, PAID, FAILED, REFUNDED)
- provider_payment_id
- provider_metadata
- created_at
- updated_at

## Эндпоинты

### 1) GET /health

Для чего: проверить, что API жив и может ходить в базу.

Успех:

- 200 OK
- Поля ответа: status, version

### 2) GET /slots

Для чего: список доступных слотов для каталога.

Query-параметры (все опциональные):

- sport
- district
- date_from
- date_to

Логика:

- По умолчанию возвращать только status = OPEN.
- Сортировка по starts_at по возрастанию (ближайшие сверху).

Успех:

- 200 OK
- Ответ: массив слотов

Ошибки:

- 422 VALIDATION_ERROR, если сломан формат даты

### 3) GET /slots/{slotId}

Для чего: детальная страница конкретного слота.

Path-параметры:

- slotId (UUID)

Успех:

- 200 OK
- Ответ: объект слота
- Желательно добавить current_participants и free_spots

Ошибки:

- 404 SLOT_NOT_FOUND

### 4) POST /slots/{slotId}/join

Для чего: пользователь бронирует место в слоте.

Требует:

- Заголовок X-Demo-Token

Path-параметры:

- slotId (UUID)

Успех:

- 201 Created
- Ответ: объект участия со статусом RESERVED

Ошибки:

- 401 UNAUTHORIZED
- 404 SLOT_NOT_FOUND
- 409 SLOT_FULL
- 409 ALREADY_JOINED
- 409 DEADLINE_PASSED

### 5) POST /slots/{slotId}/pay

Для чего: пользователь оплачивает участие.

Требует:

- Заголовок X-Demo-Token
- Заголовок X-Idempotency-Key

Path-параметры:

- slotId (UUID)

Успех:

- 200 OK
- Ответ: объект платежа со статусом PAID
- Плюс обновленное участие со статусом PAID

Ошибки:

- 401 UNAUTHORIZED
- 404 PARTICIPATION_NOT_FOUND
- 409 ALREADY_PAID
- 409 PAYMENT_IN_PROGRESS

### 6) GET /me/participations

Для чего: экран "Мои участия".

Требует:

- Заголовок X-Demo-Token

Успех:

- 200 OK
- Ответ: массив участий пользователя
- Каждый элемент должен содержать данные участия плюс краткие данные слота:
- slot_id, sport, district, venue_name, address, starts_at, deadline_at, duration_minutes, expected_price, max_price, slot_status

### 7) POST /admin/slots

Для чего: создать новый слот из админки.

Требует:

- Заголовок X-Admin-Token

Body (обязательные поля):

- sport
- district
- venue_name
- address
- starts_at
- deadline_at
- duration_minutes
- capacity
- min_players
- expected_price
- max_price

Body (опционально):

- rules_text
- status (по умолчанию OPEN)

Успех:

- 201 Created
- Ответ: созданный объект слота

Ошибки:

- 401 UNAUTHORIZED
- 422 VALIDATION_ERROR

## Что уже есть в базе и можно использовать для демо

- В seed-данных 5 слотов в статусе OPEN.
- Есть 3 демо-пользователя.
- Есть примеры участия в статусах RESERVED и PAID.
- Есть пример платежа PAID.

Это позволяет фронтенду сразу собирать экраны списка, деталей и личного кабинета на реальных данных.

## Что нужно реализовать в бэкенде по этому контракту

1. Включить маршруты /slots, /slots/{slotId}, /slots/{slotId}/join, /slots/{slotId}/pay, /me/participations, /admin/slots.
2. Добавить middleware аутентификации по X-Demo-Token и X-Admin-Token.
3. Зафиксировать единый JSON-формат ошибок.
4. Для join сделать защиту от овербукинга в транзакции.
5. Для pay сделать идемпотентность через X-Idempotency-Key.

Когда это будет реализовано, фронтенд сможет без переделок перейти с моков на живой API.


