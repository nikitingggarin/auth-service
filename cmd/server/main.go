package main

import (
	"context"
	"log"

	"auth-service/internal/auth"
	"auth-service/internal/config"
	"auth-service/internal/handler"
	"auth-service/internal/middleware"
	"auth-service/internal/repository/postgres"
	"auth-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

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
	log.Printf("Config loaded - DBURL: %s", cfg.DBURL)

	// Подключение к PostgreSQL
	ctx := context.Background()
	dbPool, err := postgres.NewPool(ctx, &postgres.Config{
		URL: cfg.DBURL,
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
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
