package locking

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	// ErrLockNotAcquired возвращается если не удалось получить блокировку
	ErrLockNotAcquired = errors.New("failed to acquire lock")

	// ErrLockNotHeld возвращается при попытке освободить чужую блокировку
	ErrLockNotHeld = errors.New("lock not held by this instance")
)

// Locker - интерфейс для distributed locking
// Позволяет легко mock'ать в тестах или менять реализацию (Redis → etcd)
type Locker interface {
	// Acquire получает блокировку с указанным TTL
	// Блокирует выполнение до получения lock или истечения context
	Acquire(ctx context.Context, key string, ttl time.Duration) (*Lock, error)

	// TryAcquire пытается получить блокировку без ожидания
	// Возвращает ErrLockNotAcquired если блокировка занята
	TryAcquire(ctx context.Context, key string, ttl time.Duration) (*Lock, error)
}

// Lock представляет полученную блокировку
type Lock struct {
	client *redis.Client
	key    string
	value  string // UUID для проверки владения
	ttl    time.Duration
}

// Release освобождает блокировку
// КРИТИЧНО: проверяет что освобождаем свою блокировку (по UUID)
func (l *Lock) Release(ctx context.Context) error {
	// Lua скрипт для атомарного сравнения и удаления
	// Защищает от случайного освобождения чужой блокировки
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Int()
	if err != nil {
		return fmt.Errorf("redis eval failed: %w", err)
	}

	if result == 0 {
		return ErrLockNotHeld
	}

	return nil
}

// Refresh продлевает TTL блокировки (для длинных операций)
func (l *Lock) Refresh(ctx context.Context, ttl time.Duration) error {
	// Проверяем что блокировка все еще наша
	currentValue, err := l.client.Get(ctx, l.key).Result()
	if err == redis.Nil {
		return ErrLockNotHeld
	}
	if err != nil {
		return fmt.Errorf("redis get failed: %w", err)
	}

	if currentValue != l.value {
		return ErrLockNotHeld
	}

	// Продлеваем TTL
	if err := l.client.Expire(ctx, l.key, ttl).Err(); err != nil {
		return fmt.Errorf("redis expire failed: %w", err)
	}

	l.ttl = ttl
	return nil
}

// TTL возвращает оставшееся время жизни блокировки
func (l *Lock) TTL(ctx context.Context) (time.Duration, error) {
	ttl, err := l.client.TTL(ctx, l.key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis ttl failed: %w", err)
	}
	return ttl, nil
}

// RedisLocker реализует Locker через Redis
type RedisLocker struct {
	client *redis.Client
}

// NewRedisLocker создает новый Redis-based locker
func NewRedisLocker(client *redis.Client) *RedisLocker {
	return &RedisLocker{
		client: client,
	}
}

// Acquire получает блокировку с ожиданием
// Использует polling с exponential backoff
func (r *RedisLocker) Acquire(ctx context.Context, key string, ttl time.Duration) (*Lock, error) {
	// Уникальный идентификатор этой блокировки
	lockValue := uuid.New().String()

	// Exponential backoff для retry
	backoff := 10 * time.Millisecond
	maxBackoff := 500 * time.Millisecond

	for {
		// Пытаемся получить блокировку
		// SetNX = SET if Not eXists (атомарная операция!)
		success, err := r.client.SetNX(ctx, key, lockValue, ttl).Result()
		if err != nil {
			return nil, fmt.Errorf("redis setnx failed: %w", err)
		}

		if success {
			// Блокировка получена!
			return &Lock{
				client: r.client,
				key:    key,
				value:  lockValue,
				ttl:    ttl,
			}, nil
		}

		// Блокировка занята, ждем с exponential backoff
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrLockNotAcquired, ctx.Err())
		case <-time.After(backoff):
			// Увеличиваем backoff (но не больше maxBackoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

// TryAcquire пытается получить блокировку без ожидания
// Возвращает ошибку если блокировка занята
func (r *RedisLocker) TryAcquire(ctx context.Context, key string, ttl time.Duration) (*Lock, error) {
	lockValue := uuid.New().String()

	// Пытаемся один раз
	success, err := r.client.SetNX(ctx, key, lockValue, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("redis setnx failed: %w", err)
	}

	if !success {
		return nil, ErrLockNotAcquired
	}

	return &Lock{
		client: r.client,
		key:    key,
		value:  lockValue,
		ttl:    ttl,
	}, nil
}

// WithLock выполняет функцию под блокировкой и автоматически освобождает lock
func WithLock(ctx context.Context, locker Locker, key string, ttl time.Duration, fn func(context.Context) error) error {
	lock, err := locker.Acquire(ctx, key, ttl)
	if err != nil {
		return err
	}

	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if releaseErr := lock.Release(releaseCtx); releaseErr != nil {
			fmt.Printf("Warning: failed to release lock %s: %v\n", key, releaseErr)
		}
	}()

	return fn(ctx)
}
