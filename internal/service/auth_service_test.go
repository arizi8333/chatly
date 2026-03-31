package service

import (
	"context"
	"testing"
	"time"

	"chatly/internal/config"
	"chatly/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// Mocks
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, u *domain.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.User), args.Error(1)
}

type MockTokenRepo struct {
	mock.Mock
}

func (m *MockTokenRepo) Create(ctx context.Context, t *domain.RefreshToken) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *MockTokenRepo) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *MockTokenRepo) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockTokenRepo) RevokeByID(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestAuthService_Register(t *testing.T) {
	userRepo := new(MockUserRepo)
	tokenRepo := new(MockTokenRepo)
	cfg := &config.Config{JWTSecret: "secret"}
	svc := NewAuthService(userRepo, tokenRepo, cfg)

	ctx := context.Background()
	username := "testuser"
	email := "test@example.com"
	password := "password123"

	// Mocking Create – it will be called with a user object containing the hashed password
	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	user, err := svc.Register(ctx, username, email, password)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, email, user.Email)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Login(t *testing.T) {
	userRepo := new(MockUserRepo)
	tokenRepo := new(MockTokenRepo)
	cfg := &config.Config{
		JWTSecret:     "secret",
		JWTExpiry:     time.Minute,
		RefreshExpiry: time.Hour,
	}
	svc := NewAuthService(userRepo, tokenRepo, cfg)

	ctx := context.Background()
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}

	userRepo.On("FindByEmail", ctx, user.Email).Return(user, nil)
	tokenRepo.On("Create", ctx, mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	access, refresh, err := svc.Login(ctx, user.Email, password)

	assert.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)
	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

func TestAuthService_Refresh_ReuseDetection(t *testing.T) {
	userRepo := new(MockUserRepo)
	tokenRepo := new(MockTokenRepo)
	cfg := &config.Config{JWTSecret: "secret"}
	svc := NewAuthService(userRepo, tokenRepo, cfg)

	ctx := context.Background()
	oldToken := "old-token"

	// Mock a revoked token (reuse attempt)
	revokedToken := &domain.RefreshToken{
		ID:      uuid.New(),
		UserID:  uuid.New(),
		Revoked: true,
	}

	tokenRepo.On("FindByHash", ctx, mock.Anything).Return(revokedToken, nil)
	tokenRepo.On("RevokeByUserID", ctx, revokedToken.UserID).Return(nil)

	_, _, err := svc.Refresh(ctx, oldToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid refresh token")
	tokenRepo.AssertExpectations(t)
}
