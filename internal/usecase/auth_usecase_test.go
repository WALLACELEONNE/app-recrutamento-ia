package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/username/app-recrutamento-ia/internal/domain"
	"github.com/username/app-recrutamento-ia/internal/usecase"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestAuthenticate_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := usecase.NewAuthUseCase(mockRepo)

	password := "mypassword"
	hashBytes, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	user := &domain.User{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		Email:          "test@example.com",
		PasswordHash:   string(hashBytes),
		Role:           "admin",
	}

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)

	token, err := uc.Authenticate(context.Background(), "test@example.com", password)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockRepo.AssertExpectations(t)
}

func TestAuthenticate_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := usecase.NewAuthUseCase(mockRepo)

	mockRepo.On("GetByEmail", mock.Anything, "unknown@example.com").Return(nil, errors.New("not found"))

	token, err := uc.Authenticate(context.Background(), "unknown@example.com", "anypassword")

	assert.Error(t, err)
	assert.Equal(t, usecase.ErrInvalidCredentials, err)
	assert.Empty(t, token)
	mockRepo.AssertExpectations(t)
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := usecase.NewAuthUseCase(mockRepo)

	password := "mypassword"
	hashBytes, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	user := &domain.User{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		Email:          "test@example.com",
		PasswordHash:   string(hashBytes),
		Role:           "admin",
	}

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)

	token, err := uc.Authenticate(context.Background(), "test@example.com", "wrongpassword")

	assert.Error(t, err)
	assert.Equal(t, usecase.ErrInvalidCredentials, err)
	assert.Empty(t, token)
	mockRepo.AssertExpectations(t)
}
