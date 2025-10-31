package service

import (
	"context"
	"errors"
	"log"
	"time"

	"auth-service/internal/cache"
	"auth-service/internal/email"
	"auth-service/internal/models"
	"auth-service/internal/repository/postgres"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo     UserRepository
	jwtService   JWTService
	userCache    *cache.UserCache
	emailService *email.EmailService
}

func NewAuthService(userRepo UserRepository, jwtService JWTService) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		jwtService:   jwtService,
		userCache:    cache.NewUserCache(5 * time.Minute),
		emailService: email.NewEmailService(),
	}
}

type AuthResponse struct {
	User  *models.User `json:"user"`
	Token string       `json:"token"`
}

// GetUserByEmail с кешированием
func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	// Пытаемся получить из кеша
	if cachedUser := s.userCache.Get(email); cachedUser != nil {
		return cachedUser, nil
	}

	// Если нет в кеше - идем в БД
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Сохраняем в кеш
	if user != nil {
		s.userCache.Set(email, user)
	}

	return user, nil
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

	// Инвалидируем кеш при создании нового пользователя
	s.userCache.Delete(req.Email)

	// Генерируем JWT токен
	token, err := s.jwtService.GenerateToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	// 🔥 ЗАПУСКАЕМ ФОНОВУЮ ОТПРАВКУ EMAIL
	s.emailService.SendWelcomeEmailAsync(user.Email, user.Email)

	log.Printf("🚀 Welcome email sending started in background for: %s", user.Email)

	return &AuthResponse{
		User:  user,
		Token: token,
	}, nil
}

// Login выполняет вход пользователя
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*AuthResponse, error) {
	// Используем кешированный метод
	user, err := s.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
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

	return user, nil
}
