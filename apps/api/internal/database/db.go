package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations выполняет все pending миграции при старте приложения
// КРИТИЧНО: Вызывается перед запуском HTTP-сервера
func RunMigrations(databaseURL string) error {
	m, err := migrate.New(
		"file://internal/database/migrations",
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Выполняем миграции до последней версии
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	// Проверяем dirty state (миграция упала на половине)
	if dirty {
		log.Printf("WARNING: Database is in dirty state at version %d", version)
		return fmt.Errorf("database is in dirty state")
	}

	log.Printf("✓ Migrations completed successfully. Current version: %d", version)
	return nil
}

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}
	config.MaxConns = 25                        // Максимум открытых соединений
	config.MinConns = 5                         // Минимум idle соединений
	config.MaxConnLifetime = 0                  // Соединения живут бесконечно
	config.MaxConnIdleTime = 0                  // Idle соединения не закрываются
	config.HealthCheckPeriod = 30 * time.Second // Периодический health check (должен быть > 0)

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("✓ Database connection established (max_conns=%d, min_conns=%d)",
		config.MaxConns, config.MinConns)

	return pool, nil
}
