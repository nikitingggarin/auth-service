package service

import (
	"context"

	"auth-service/internal/auth"
	"auth-service/internal/models"
)

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	CreateUser(ctx context.Context, req *models.CreateUserRequest, passwordHash string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	UserExists(ctx context.Context, email string) (bool, error)
}

// JWTService интерфейс для работы с JWT токенами
type JWTService interface {
	GenerateToken(userID, email string) (string, error)
	ValidateToken(tokenString string) (*auth.Claims, error)
}
