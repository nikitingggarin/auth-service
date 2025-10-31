package service

import (
	"context"
	"testing"

	"auth-service/internal/auth"
	"auth-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository - мок репозитория пользователей
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, req *models.CreateUserRequest, passwordHash string) (*models.User, error) {
	args := m.Called(ctx, req, passwordHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UserExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// MockJWTService - мок JWT сервиса
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(userID, email string) (string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) ValidateToken(tokenString string) (*auth.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

func TestAuthService_Register_Success(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockJWTService := new(MockJWTService)
	authService := NewAuthService(mockUserRepo, mockJWTService)

	req := &models.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Настраиваем моки
	mockUserRepo.On("UserExists", mock.Anything, req.Email).Return(false, nil)
	mockUserRepo.On("CreateUser", mock.Anything, req, mock.AnythingOfType("string")).Return(&models.User{
		ID:    [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Email: req.Email,
	}, nil)
	mockJWTService.On("GenerateToken", mock.Anything, req.Email).Return("jwt-token", nil)

	// Act
	result, err := authService.Register(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test@example.com", result.User.Email)
	assert.Equal(t, "jwt-token", result.Token)

	mockUserRepo.AssertExpectations(t)
	mockJWTService.AssertExpectations(t)
}

func TestAuthService_Register_UserExists(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockJWTService := new(MockJWTService)
	authService := NewAuthService(mockUserRepo, mockJWTService)

	req := &models.CreateUserRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	// Настраиваем моки
	mockUserRepo.On("UserExists", mock.Anything, req.Email).Return(true, nil)

	// Act
	result, err := authService.Register(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "user with this email already exists", err.Error())

	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockJWTService := new(MockJWTService)
	authService := NewAuthService(mockUserRepo, mockJWTService)

	req := &models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Создаем реальный bcrypt хеш для теста
	realPasswordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to generate password hash: %v", err)
	}

	user := &models.User{
		ID:           [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Email:        "test@example.com",
		PasswordHash: string(realPasswordHash),
	}

	// Настраиваем моки
	mockUserRepo.On("GetUserByEmail", mock.Anything, req.Email).Return(user, nil)
	mockJWTService.On("GenerateToken", mock.Anything, req.Email).Return("jwt-token", nil)

	// Act
	result, err := authService.Login(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test@example.com", result.User.Email)
	assert.Equal(t, "jwt-token", result.Token)

	mockUserRepo.AssertExpectations(t)
	mockJWTService.AssertExpectations(t)
}

func TestAuthService_GetProfile_Success(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockJWTService := new(MockJWTService)
	authService := NewAuthService(mockUserRepo, mockJWTService)

	userID := "test-user-id"
	user := &models.User{
		ID:    [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Email: "test@example.com",
	}

	// Настраиваем моки
	mockUserRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)

	// Act
	result, err := authService.GetProfile(context.Background(), userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test@example.com", result.Email)

	mockUserRepo.AssertExpectations(t)
}
