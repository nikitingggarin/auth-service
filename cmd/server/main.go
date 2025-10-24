package main

import (
	"context"
	"log"
	"net/url"

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
	// Загрузка .env файла с явным указанием пути
	if err := godotenv.Load("C:/projects_Go/auth-service/.env"); err != nil {
		log.Printf("Failed to load .env file: %v", err)
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

	log.Printf("🚀 Server starting on :%s...", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
