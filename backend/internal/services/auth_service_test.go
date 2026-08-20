package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"clothing-store-api/internal/models"
	"clothing-store-api/internal/repositories"
)

type memoryUserRepository struct{ users map[string]models.User }

func newMemoryUserRepository() *memoryUserRepository {
	return &memoryUserRepository{users: map[string]models.User{}}
}
func (r *memoryUserRepository) Create(_ context.Context, u models.User) (models.User, error) {
	if _, exists := r.users[u.Email]; exists {
		return models.User{}, repositories.ErrEmailAlreadyExists
	}
	r.users[u.Email] = u
	return u, nil
}
func (r *memoryUserRepository) FindByEmail(_ context.Context, email string) (models.User, error) {
	u, exists := r.users[email]
	if !exists {
		return models.User{}, repositories.ErrUserNotFound
	}
	return u, nil
}

func TestAuthServiceRegisterAndLogin(t *testing.T) {
	repo := newMemoryUserRepository()
	service := NewAuthService(repo, "test-secret")
	service.now = func() time.Time { return time.Unix(2_000_000_000, 0) }

	created, err := service.Register(context.Background(), RegisterInput{Name: "Test User", Email: "TEST@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if created.Email != "test@example.com" || created.Role != DefaultRole || bcrypt.CompareHashAndPassword([]byte(created.PasswordHash), []byte("password123")) != nil {
		t.Fatalf("unexpected registered user: %+v", created)
	}
	token, loggedIn, err := service.Login(context.Background(), "test@example.com", "password123")
	if err != nil || loggedIn.ID != created.ID {
		t.Fatalf("Login() error = %v, user = %+v", err, loggedIn)
	}
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) { return []byte("test-secret"), nil })
	if err != nil || !parsed.Valid {
		t.Fatalf("JWT invalid: %v", err)
	}
}

func TestAuthServiceValidationDuplicateAndInvalidCredentials(t *testing.T) {
	repo := newMemoryUserRepository()
	service := NewAuthService(repo, "test-secret")
	if _, err := service.Register(context.Background(), RegisterInput{Name: "x", Email: "bad", Password: "short"}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("validation error = %v", err)
	}
	input := RegisterInput{Name: "Test User", Email: "test@example.com", Password: "password123"}
	if _, err := service.Register(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Register(context.Background(), input); !errors.Is(err, repositories.ErrEmailAlreadyExists) {
		t.Fatalf("duplicate error = %v", err)
	}
	if _, _, err := service.Login(context.Background(), "test@example.com", "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("credentials error = %v", err)
	}
	if _, _, err := service.Login(context.Background(), "missing@example.com", "password123"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("missing user error = %v", err)
	}
}
