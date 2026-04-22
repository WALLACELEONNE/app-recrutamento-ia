package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/username/app-recrutamento-ia/internal/auth"
)

func TestGenerateAndValidateToken(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()
	role := "admin"

	// Test GenerateToken
	tokenString, err := auth.GenerateToken(userID, orgID, role)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Test ValidateToken
	claims, err := auth.ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, claims)

	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, orgID, claims.OrganizationID)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, "nova-voice-auth", claims.Issuer)

	// Test Expired Token logic (manual creation)
	// We can't easily manipulate time.Now() without a mock clock,
	// so we'll create an expired token manually to test validation failure.
	expiredClaims := &auth.Claims{
		UserID:         userID,
		OrganizationID: orgID,
		Role:           role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "nova-voice-auth",
		},
	}

	// This uses the same secret as auth.go, we know the default is "super-secret-key-for-local-dev-only"
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, _ := expiredToken.SignedString([]byte("super-secret-key-for-local-dev-only"))

	_, err = auth.ValidateToken(expiredTokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()
	role := "admin"

	claims := &auth.Claims{
		UserID:         userID,
		OrganizationID: orgID,
		Role:           role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "nova-voice-auth",
		},
	}

	// Sign with a different key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	invalidTokenString, _ := token.SignedString([]byte("wrong-secret-key"))

	_, err := auth.ValidateToken(invalidTokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")
}
