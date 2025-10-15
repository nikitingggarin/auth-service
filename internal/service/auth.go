package service

import (
	"context"
	"errors"

	"auth-service/internal/auth"
	"auth-service/internal/models"
	"auth-service/internal/repository/postgres"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo   *postgres.UserRepository
	jwtService *auth.JWTService
}

func NewAuthService(userRepo *postgres.UserRepository, jwtService *auth.JWTService) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

type AuthResponse struct {
	User  *models.User `json:"user"`
	Token string       `json:"token"`
}

// hashPassword хеширует пароль
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// checkPassword проверяет пароль
func checkPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, req *models.CreateUserRequest) (*AuthResponse, error) {
	// Проверяем что пользователь с таким email не существует
	exists, err := s.userRepo.UserExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user with this email already exists")
	}

	// Хешируем пароль
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Создаем пользователя
	user, err := s.userRepo.CreateUser(ctx, req, passwordHash)
	if err != nil {
		return nil, err
	}

	// Генерируем JWT токен
	token, err := s.jwtService.GenerateToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &AuthResponse{
		User:  user,
		Token: token,
	}, nil
}

// Login выполняет вход пользователя
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*AuthResponse, error) {
	// Получаем пользователя по email
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	// Проверяем пароль
	if err := checkPassword(user.PasswordHash, req.Password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Генерируем JWT токен
	token, err := s.jwtService.GenerateToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &AuthResponse{
		User:  user,
		Token: token,
	}, nil
}

// GetProfile получает профиль пользователя по ID
func (s *AuthService) GetProfile(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}
