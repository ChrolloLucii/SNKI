# Внедрение Middleware Авторизации

В рамках второго этапа MVP завершена реализация проверки токенов через Chi middleware и рефакторинг обработчиков на использование контекста запроса. 

## Что было сделано:
1. **Создание Middleware (Auth middlewares):**
   - Написаны `AuthUserMiddleware` и `AuthAdminMiddleware`. 
   - Они проверяют наличие заголовков `X-Demo-Token` и `X-Admin-Token` соответственно.
   - Если заголовок отсутствует, возвращается стандартная ошибка `401 Unauthorized`.
   - Если заголовок передан, значение токена извлекается, внедряется в `r.Context()` с помощью `context.WithValue` и запрос передается дальше по цепочке.
   - Для удобного извлечения `userID` из контекста написана хелпер-функция `GetUserID(ctx)`.

2. **Применение через `r.Group`:**
   - В точке сборки API (`apps/api/cmd/api/main.go`) роутинг переписан с использованием `r.Group`.
   - Защищенные маршруты для пользователей (`POST /slots/{slotId}/join`, `POST /slots/{slotId}/pay`, `GET /me/participations`) закрыты через `r.Use(handlers.AuthUserMiddleware)`.
   - Защищенные маршруты для администраторов (`POST /admin/slots`) закрыты через `r.Use(handlers.AuthAdminMiddleware)`.

3. **Рефакторинг API хэндлеров:**
   - Логика извлечения `user_id` из пейлоада тела запроса или заголовков внутри самих функций `Join`, `Pay`, `GetMyParticipations` удалена.
   - Теперь все ручки требуют `user_id` непосредственно из контекста (`userID := GetUserID(r.Context())`). Это исключает дублирование кода, упрощает логику и повышает безопасность (пользователь не может подделать `user_id` в JSON-теле, если оно игнорируется, и доверяем только токену в заголовке).

## Где находится код:
- Миддлвари и хелпер `GetUserID`: `apps/api/internal/handlers/middleware.go`
- Регистрация маршрутов с защитой: `apps/api/cmd/api/main.go`
- Хэндлеры слотов (Join/Pay): `apps/api/internal/handlers/routes.go`
- Хэндлеры /me: `apps/api/internal/handlers/me.go`
- Хэндлеры администратора: `apps/api/internal/handlers/admin.go`

## Как тестировать:

1. Ошибка авторизации пользователя (отсутствует `X-Demo-Token`):
```bash
curl -i http://localhost:8080/me/participations
```
*Ожидаемый ответ: `401 Unauthorized` с JSON ошибкой.*

2. Успешный доступ пользователя:
```bash
curl -i http://localhost:8080/me/participations -H "X-Demo-Token: user-123"
```
*Ожидаемый ответ: `200 OK` (вероятно `[]`, если участия нет).*

3. Ошибка авторизации администратора (отсутствует `X-Admin-Token`):
```bash
curl -i -X POST http://localhost:8080/admin/slots
```

4. Успешный доступ к админ панеле:
```bash
curl -i -X POST http://localhost:8080/admin/slots -H "X-Admin-Token: supersecret" -H "Content-Type: application/json" -d '{"name": "test slot"}'
```
*Ожидаемый ответ: `201 Created` / `200 OK` (в зависимости от формата тела).*
