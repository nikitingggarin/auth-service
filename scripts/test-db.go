package main

import (
	"context"
	"log"
	"os"

	"auth-service/internal/repository/postgres"
	"github.com/joho/godotenv"
)

func main() {
	// Загрузка .env файла из корня проекта
	if err := godotenv.Load("C:/projects_Go/auth-service/.env"); err != nil {
		log.Fatal("Failed to load .env file:", err)
	}

	dbURL := os.Getenv("DB_URL")
	log.Printf("Testing connection to: %s", dbURL)

	// Подключение к PostgreSQL
	ctx := context.Background()
	dbPool, err := postgres.NewPool(ctx, &postgres.Config{
		URL: dbURL,
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer dbPool.Close()

	log.Println("✅ Successfully connected to PostgreSQL!")

	// Проверим что таблица users доступна
	var tableName string
	err = dbPool.QueryRow(ctx, "SELECT table_name FROM information_schema.tables WHERE table_name = 'users'").Scan(&tableName)
	if err != nil {
		log.Fatal("Table 'users' not found:", err)
	}

	log.Printf("✅ Table '%s' exists!", tableName)
}
