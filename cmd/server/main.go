package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-service/internal/auth"
	"auth-service/internal/config"
	"auth-service/internal/handler"
	"auth-service/internal/middleware"
	"auth-service/internal/repository/postgres"
	"auth-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// maskDBURL маскирует пароль в URL базы данных
func maskDBURL(dbURL string) string {
	parsed, err := url.Parse(dbURL)
	if err != nil {
		return "invalid-db-url"
	}

	// Если есть пароль - маскируем его
	if parsed.User != nil {
		if _, hasPassword := parsed.User.Password(); hasPassword {
			// Заменяем пароль на ****
			parsed.User = url.UserPassword(parsed.User.Username(), "****")
		}
	}

	return parsed.String()
}

func main() {
	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ Failed to load .env file: %v", err)
		log.Println("Using default configuration")
	} else {
		log.Println("✅ .env file loaded successfully")
	}

	// Загрузка конфигурации
	cfg := config.Load()

	// Логируем маскированный URL для безопасности
	maskedDBURL := maskDBURL(cfg.DBURL)
	log.Printf("Config loaded - DBURL: %s", maskedDBURL)

	// Подключение к PostgreSQL
	ctx := context.Background()
	dbPool, err := postgres.NewPool(ctx, &postgres.Config{
		URL: cfg.DBURL,
	})
	if err != nil {
		log.Fatal("Failed to connect to database")
	}
	defer dbPool.Close()

	log.Println("✅ Database connection established")

	// Инициализация JWT сервиса
	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTExpiration)

	// Инициализация репозиториев и сервисов
	userRepo := postgres.NewUserRepository(dbPool)
	authService := service.NewAuthService(userRepo, jwtService)
	authHandler := handler.NewAuthHandler(authService)

	// Создание Gin роутера
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "OK",
			"message": "Auth service is running!",
			"port":    cfg.ServerPort,
		})
	})

	// Concurrency test endpoint
	r.GET("/concurrent", func(c *gin.Context) {
		// Симулируем тяжелую операцию (например, сложный запрос к БД)
		time.Sleep(2 * time.Second)

		c.JSON(200, gin.H{
			"message":      "Concurrent request processed",
			"time":         time.Now().Format(time.RFC3339),
			"goroutine_id": getGoroutineID(), // Покажем ID горутины
		})
	})

	// Auth routes
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
	}

	// Protected routes (require JWT token)
	protectedGroup := r.Group("/api")
	protectedGroup.Use(middleware.AuthMiddleware(jwtService))
	{
		protectedGroup.GET("/profile", authHandler.GetProfile)
	}

	// Создаем HTTP сервер с настройками
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	// Запускаем сервер в отдельной goroutine
	go func() {
		log.Printf("🚀 Server starting on :%s...", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Канал для сигналов OS
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Ждем сигнал завершения
	<-quit
	log.Println("🛑 Shutting down server...")

	// Graceful shutdown с таймаутом 30 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited properly")
}