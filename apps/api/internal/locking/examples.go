package locking

import (
	"context"
	"fmt"
	"time"
)

// ============================================================================
// ПРИМЕРЫ ИСПОЛЬЗОВАНИЯ Redis Distributed Locking
// ============================================================================

// Пример 1: JOIN к слоту с защитой от овербукинга
func ExampleJoinSlotWithLock(locker Locker, slotID, userID string) error {
	ctx := context.Background()

	// Ключ блокировки специфичный для слота
	lockKey := fmt.Sprintf("slot:%s:join", slotID)

	// TTL = 5 секунд (достаточно для проверки capacity + INSERT)
	err := WithLock(ctx, locker, lockKey, 5*time.Second, func(ctx context.Context) error {
		// ===== КРИТИЧЕСКАЯ СЕКЦИЯ (под блокировкой) =====

		// 1. Начинаем транзакцию PostgreSQL
		// tx, _ := db.BeginTx(ctx, ...)

		// 2. Проверяем capacity
		// count, _ := tx.QueryRow("SELECT COUNT(*) FROM participants WHERE slot_id = $1", slotID)
		// slot, _ := tx.QueryRow("SELECT capacity FROM slots WHERE id = $1", slotID)

		// if count >= capacity {
		//     return ErrSlotFull
		// }

		// 3. Вставляем участника
		// tx.Exec("INSERT INTO participants (slot_id, user_id, status) VALUES ($1, $2, 'RESERVED')", slotID, userID)

		// 4. Коммитим
		// tx.Commit()

		fmt.Printf("✓ User %s joined slot %s\n", userID, slotID)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to join slot: %w", err)
	}

	return nil
}

// Пример 2: Оплата с idempotency key и lock
func ExamplePayWithIdempotencyAndLock(locker Locker, participantID, idempotencyKey string) error {
	ctx := context.Background()

	// Двойная защита:
	// 1. Idempotency key в БД (UNIQUE constraint)
	// 2. Redis lock на время обработки платежа

	lockKey := fmt.Sprintf("payment:participant:%s", participantID)

	// TTL = 30 секунд (платежи могут быть долгими: вызов Stripe API)
	err := WithLock(ctx, locker, lockKey, 30*time.Second, func(ctx context.Context) error {
		// ===== КРИТИЧЕСКАЯ СЕКЦИЯ =====

		// 1. Проверяем idempotency key
		// existingPayment, _ := queries.GetPaymentByIdempotencyKey(ctx, idempotencyKey)
		// if existingPayment != nil {
		//     // Платеж уже существует, возвращаем cached response
		//     return existingPayment
		// }

		// 2. Проверяем статус участника
		// participant, _ := queries.GetParticipant(ctx, participantID)
		// if participant.Status == "PAID" {
		//     return ErrAlreadyPaid
		// }

		// 3. Создаем платеж в БД
		// payment, _ := queries.CreatePayment(ctx, CreatePaymentParams{
		//     ParticipantID:   participantID,
		//     IdempotencyKey:  idempotencyKey,
		//     Amount:          500,
		//     Currency:        "RUB",
		//     Provider:        "FAKE",
		//     Status:          "PAID",
		// })

		// 4. Обновляем статус участника
		// queries.UpdateParticipantStatus(ctx, participantID, "PAID")

		fmt.Printf("✓ Payment processed for participant %s (idempotency: %s)\n", participantID, idempotencyKey)
		return nil
	})

	return err
}

// Пример 3: Retry при неудаче получения блокировки
func ExampleAcquireWithRetry(locker Locker) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lockKey := "slot:123:join"

	// Acquire будет пытаться получить блокировку в течение 10 секунд
	// с exponential backoff (10ms → 20ms → 40ms → ... → 500ms)
	lock, err := locker.Acquire(ctx, lockKey, 5*time.Second)
	if err != nil {
		// Не удалось получить блокировку за 10 секунд
		// Либо context timeout, либо Redis недоступен
		fmt.Printf("Failed to acquire lock: %v\n", err)
		return
	}

	// Блокировка получена!
	defer lock.Release(context.Background())

	// Выполняем критическую секцию
	fmt.Println("Working under lock...")
	time.Sleep(2 * time.Second)
	fmt.Println("Done!")
}

// Пример 4: TryAcquire (без ожидания)
func ExampleTryAcquireWithoutWait(locker Locker) {
	ctx := context.Background()
	lockKey := "slot:123:join"

	// Пытаемся получить блокировку БЕЗ ОЖИДАНИЯ
	// Если занята - сразу возвращаем ошибку
	lock, err := locker.TryAcquire(ctx, lockKey, 5*time.Second)
	if err == ErrLockNotAcquired {
		// Блокировка занята, возвращаем HTTP 429 Too Many Requests
		fmt.Println("Slot is being processed by another request, try again later")
		return
	}
	if err != nil {
		fmt.Printf("Redis error: %v\n", err)
		return
	}

	defer lock.Release(ctx)

	// Блокировка получена, выполняем работу
	fmt.Println("Processing...")
}

// Пример 5: Refresh (продление TTL для длинных операций)
func ExampleRefreshLock(locker Locker) {
	ctx := context.Background()
	lockKey := "long-task:123"

	// Получаем блокировку на 10 секунд
	lock, err := locker.Acquire(ctx, lockKey, 10*time.Second)
	if err != nil {
		fmt.Printf("Failed to acquire lock: %v\n", err)
		return
	}
	defer lock.Release(ctx)

	// Начинаем длинную операцию (например, обработка большого файла)
	for i := 0; i < 5; i++ {
		fmt.Printf("Step %d of long operation...\n", i+1)
		time.Sleep(5 * time.Second)

		// Продлеваем блокировку на еще 10 секунд
		if err := lock.Refresh(ctx, 10*time.Second); err != nil {
			fmt.Printf("Failed to refresh lock: %v\n", err)
			return
		}
		fmt.Println("✓ Lock refreshed")
	}

	fmt.Println("Long operation completed!")
}

// Пример 6: Проверка TTL блокировки
func ExampleCheckLockTTL(locker Locker) {
	ctx := context.Background()
	lockKey := "task:123"

	lock, _ := locker.Acquire(ctx, lockKey, 30*time.Second)
	defer lock.Release(ctx)

	// Проверяем сколько времени осталось до истечения блокировки
	ttl, err := lock.TTL(ctx)
	if err != nil {
		fmt.Printf("Failed to get TTL: %v\n", err)
		return
	}

	fmt.Printf("Lock will expire in %v\n", ttl)

	// Если осталось мало времени - продлеваем
	if ttl < 5*time.Second {
		lock.Refresh(ctx, 30*time.Second)
		fmt.Println("Lock refreshed")
	}
}
