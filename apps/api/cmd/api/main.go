package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ChrolloLucii/SNKI/apps/api/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	// "github.com/ChrolloLucii/SNKI/apps/api/internal/locking"
	// "github.com/redis/go-redis/v9"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println(" Starting Zal API...")

	// 1. CONFIGURATION

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Redis временно отключен.
	// redisURL := os.Getenv("REDIS_URL")
	// if redisURL == "" {
	// 	log.Fatal("REDIS_URL environment variable is required")
	// }

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	// 2. DATABASE SETUP

	ctx := context.Background()

	// Запуск миграций (КРИТИЧНО: перед созданием connection pool)
	log.Println(" Running database migrations...")
	if err := database.RunMigrations(databaseURL); err != nil {
		log.Fatalf(" Failed to run migrations: %v", err)
	}

	// Создание connection pool
	log.Println(" Connecting to PostgreSQL...")
	pool, err := database.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf(" Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// 3. REDIS SETUP (временно отключено)
	// log.Println(" Connecting to Redis...")
	// opts, err := redis.ParseURL(redisURL)
	// if err != nil {
	// 	log.Fatalf(" Failed to parse Redis URL: %v", err)
	// }

	// redisClient := redis.NewClient(opts)
	// defer redisClient.Close()

	// // Проверка соединения с Redis
	// if err := redisClient.Ping(ctx).Err(); err != nil {
	// 	log.Fatalf(" Failed to ping Redis: %v", err)
	// }
	// log.Println(" Redis connection established")

	// // Создание locker для distributed locking
	// locker := locking.NewRedisLocker(redisClient)
	// _ = locker // TODO: Использовать в handlers

	// 4. HTTP SERVER SETUP
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)                 // Уникальный ID для каждого запроса
	r.Use(middleware.RealIP)                    // Получение реального IP клиента
	r.Use(middleware.Logger)                    // Логирование всех запросов
	r.Use(middleware.Recoverer)                 // Восстановление после panic
	r.Use(middleware.Timeout(60 * time.Second)) // Таймаут 60 сек

	// CORS для frontend
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Demo-Token", "X-Admin-Token", "X-Idempotency-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // 5 минут
	}))

	// 5. ROUTES

	// Health check (для мониторинга и k8s liveness probe)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		// Проверка PostgreSQL
		if err := pool.Ping(ctx); err != nil {
			http.Error(w, `{"status":"unhealthy","reason":"database"}`, http.StatusServiceUnavailable)
			return
		}

		// Проверка Redis (временно отключена)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok","version":"0.1.0"}`)
	})

	// TODO: API endpoints
	// r.Get("/slots", handlers.ListSlots(pool))
	// r.Get("/slots/{slotId}", handlers.GetSlot(pool))
	// r.Post("/slots/{slotId}/join", handlers.JoinSlot(pool, locker))
	// r.Post("/slots/{slotId}/pay", handlers.PaySlot(pool, locker))
	// r.Get("/me/participations", handlers.GetMyParticipations(pool))
	// r.Post("/admin/slots", handlers.CreateSlot(pool))

	// 404 handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":"Not Found"}`)
	})
	// 6. START SERVER
	server := &http.Server{
		Addr:           ":" + port,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println(" Shutting down gracefully...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf(" Server shutdown error: %v", err)
		}
	}()

	log.Printf("✓ Server starting on port %s (env=%s)", port, env)
	log.Printf("  Health: http://localhost:%s/health", port)
	log.Println("─────────────────────────────────────────────────")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf(" Server failed: %v", err)
	}

	log.Println(" Server stopped")
}
