package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"chatly/internal/config"
	"chatly/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  domain.UserRepository
	tokenRepo domain.TokenRepository
	cfg       *config.Config
}

func NewAuthService(userRepo domain.UserRepository, tokenRepo domain.TokenRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		cfg:       cfg,
	}
}

func (s *AuthService) Register(ctx context.Context, username, email, password string) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:        uuid.New(),
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, err := s.GenerateAccessToken(user.ID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.GenerateRefreshToken(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) GenerateAccessToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(s.cfg.JWTExpiry).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) GenerateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	rawToken := uuid.New().String()
	hash := hashToken(rawToken)

	token := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(s.cfg.RefreshExpiry),
		Revoked:   false,
		CreatedAt: time.Now(),
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return "", err
	}

	return rawToken, nil
}

func (s *AuthService) Refresh(ctx context.Context, rawToken string) (string, string, error) {
	hash := hashToken(rawToken)
	token, err := s.tokenRepo.FindByHash(ctx, hash)
	if err != nil || token.Revoked || time.Now().After(token.ExpiresAt) {
		if token != nil {
			// Rotation security: if an old/revoked token is used, revoke all tokens for that user
			s.tokenRepo.RevokeByUserID(ctx, token.UserID)
		}
		return "", "", errors.New("invalid refresh token")
	}

	// Revoke current token
	s.tokenRepo.RevokeByID(ctx, token.ID)

	// Issue new tokens
	accessToken, err := s.GenerateAccessToken(token.UserID)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.GenerateRefreshToken(ctx, token.UserID)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

func hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}
