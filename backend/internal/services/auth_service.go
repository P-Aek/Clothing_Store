package services

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"clothing-store-api/internal/models"
	"clothing-store-api/internal/repositories"
	"clothing-store-api/internal/utils"
)

const DefaultRole = "customer"

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidInput       = errors.New("invalid input")
)

type AuthService struct {
	users     repositories.UserRepository
	jwtSecret string
	now       func() time.Time
}

func NewAuthService(users repositories.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret, now: time.Now}
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (models.User, error) {
	name, email, err := validateCredentials(input.Name, input.Email, input.Password)
	if err != nil {
		return models.User{}, err
	}
	hash, err := utils.HashPassword(input.Password)
	if err != nil {
		return models.User{}, fmt.Errorf("hash password: %w", err)
	}
	now := s.now().UTC()
	return s.users.Create(ctx, models.User{
		ID:           primitive.NewObjectID(),
		Name:         name,
		Email:        email,
		PasswordHash: hash,
		Role:         DefaultRole,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, models.User, error) {
	_, normalizedEmail, err := validateCredentials("valid name", email, password)
	if err != nil {
		return "", models.User{}, ErrInvalidCredentials
	}
	found, err := s.users.FindByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return "", models.User{}, ErrInvalidCredentials
		}
		return "", models.User{}, fmt.Errorf("find user: %w", err)
	}
	if !utils.CheckPassword(found.PasswordHash, password) {
		return "", models.User{}, ErrInvalidCredentials
	}
	token, err := utils.GenerateJWT(found.ID.Hex(), found.Role, s.jwtSecret, s.now())
	if err != nil {
		return "", models.User{}, fmt.Errorf("sign token: %w", err)
	}
	return token, found, nil
}

func validateCredentials(name, email, password string) (string, string, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	if len(name) < 2 || len(name) > 100 || len(password) < 8 || len(password) > 72 {
		return "", "", ErrInvalidInput
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || !strings.Contains(email, "@") {
		return "", "", ErrInvalidInput
	}
	return name, email, nil
}
