package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/username/app-recrutamento-ia/internal/auth"
	"github.com/username/app-recrutamento-ia/internal/domain"
	"github.com/username/app-recrutamento-ia/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type AuthUseCase struct {
	userRepo domain.UserRepository
}

func NewAuthUseCase(repo domain.UserRepository) *AuthUseCase {
	return &AuthUseCase{userRepo: repo}
}

// Authenticate verifies the email and password, and returns a JWT token.
func (uc *AuthUseCase) Authenticate(ctx context.Context, email, password string) (string, error) {
	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		logger.Warn("Failed login attempt: user not found", zap.String("email", email))
		return "", ErrInvalidCredentials
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		logger.Warn("Failed login attempt: invalid password", zap.String("email", email))
		return "", ErrInvalidCredentials
	}

	// Generate JWT
	token, err := auth.GenerateToken(user.ID, user.OrganizationID, user.Role)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	logger.Info("User authenticated successfully", zap.String("user_id", user.ID.String()))
	return token, nil
}

// GeneratePasswordHash is a helper utility to create new users (seeds or registration).
func GeneratePasswordHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}
