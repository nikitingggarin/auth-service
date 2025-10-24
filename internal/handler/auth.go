package handler

import (
	"context"
	"net/http"

	"auth-service/internal/models"
	"auth-service/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthService интерфейс для работы с аутентификацией
type AuthService interface {
	Register(ctx context.Context, req *models.CreateUserRequest) (*service.AuthResponse, error)
	Login(ctx context.Context, req *models.LoginRequest) (*service.AuthResponse, error)
	GetProfile(ctx context.Context, userID string) (*models.User, error)
}

type AuthHandler struct {
	authService AuthService
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register обрабатывает регистрацию пользователя
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.CreateUserRequest

	// Валидация входных данных
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Создание пользователя
	authResponse, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to register user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":         authResponse.User.ID,
			"email":      authResponse.User.Email,
			"created_at": authResponse.User.CreatedAt,
		},
		"token": authResponse.Token,
	})
}

// Login обрабатывает вход пользователя
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	// Валидация входных данных
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Аутентификация пользователя
	authResponse, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":    authResponse.User.ID,
			"email": authResponse.User.Email,
		},
		"token": authResponse.Token,
	})
}

// GetProfile возвращает профиль текущего пользователя
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Получаем user_id из контекста (устанавливается middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Получаем профиль пользователя
	user, err := h.authService.GetProfile(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Failed to get user profile",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}