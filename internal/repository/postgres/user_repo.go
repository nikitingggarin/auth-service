package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auth-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Доменные ошибки
var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser создает нового пользователя
func (r *UserRepository) CreateUser(ctx context.Context, req *models.CreateUserRequest, passwordHash string) (*models.User, error) {
	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO users (id, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByEmail получает пользователя по email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users 
		WHERE email = $1
	`

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// UserExists проверяет существует ли пользователь с таким email
func (r *UserRepository) UserExists(ctx context.Context, email string) (bool, error) {
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}

// GetUserByID получает пользователя по ID
func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User

	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users 
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}