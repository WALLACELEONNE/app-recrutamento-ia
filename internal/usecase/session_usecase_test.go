package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/username/app-recrutamento-ia/internal/domain"
	"github.com/username/app-recrutamento-ia/internal/usecase"
)

// MockSessionRepository mocks the domain.SessionRepository interface
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *domain.InterviewSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.InterviewSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.InterviewSession), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSessionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SessionStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockSessionRepository) GetDashboardMetrics(ctx context.Context) (domain.DashboardMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(domain.DashboardMetrics), args.Error(1)
}

func (m *MockSessionRepository) GetRecentInterviews(ctx context.Context, limit int) ([]domain.RecentInterview, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]domain.RecentInterview), args.Error(1)
}

func (m *MockSessionRepository) GetJobs(ctx context.Context) ([]domain.Job, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Job), args.Error(1)
}

func (m *MockSessionRepository) CreateJob(ctx context.Context, job *domain.Job) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockSessionRepository) CreateCandidate(ctx context.Context, candidate *domain.Candidate) error {
	args := m.Called(ctx, candidate)
	return args.Error(0)
}

func (m *MockSessionRepository) GetSessionReport(ctx context.Context, sessionID uuid.UUID) (domain.ReportData, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(domain.ReportData), args.Error(1)
}

func TestStartSession_InvalidID(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	uc := usecase.NewSessionUseCase(mockRepo)

	err := uc.StartSession(context.Background(), uuid.Nil)

	assert.Error(t, err)
	assert.Equal(t, "invalid session ID", err.Error())
	mockRepo.AssertNotCalled(t, "GetByID")
}

func TestStartSession_RepoGetError(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	uc := usecase.NewSessionUseCase(mockRepo)
	sessionID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, sessionID).Return(nil, errors.New("db error"))

	err := uc.StartSession(context.Background(), sessionID)

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestStartSession_AlreadyDone(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	uc := usecase.NewSessionUseCase(mockRepo)
	sessionID := uuid.New()

	session := &domain.InterviewSession{
		ID:     sessionID,
		Status: domain.SessionStatusDone,
	}

	mockRepo.On("GetByID", mock.Anything, sessionID).Return(session, nil)

	err := uc.StartSession(context.Background(), sessionID)

	assert.Error(t, err)
	assert.Equal(t, "session is already done", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestStartSession_Success(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	uc := usecase.NewSessionUseCase(mockRepo)
	sessionID := uuid.New()

	session := &domain.InterviewSession{
		ID:     sessionID,
		Status: domain.SessionStatusInvited,
	}

	mockRepo.On("GetByID", mock.Anything, sessionID).Return(session, nil)
	mockRepo.On("UpdateStatus", mock.Anything, sessionID, domain.SessionStatusInProgress).Return(nil)

	err := uc.StartSession(context.Background(), sessionID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetSessionDetails_InvalidID(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	uc := usecase.NewSessionUseCase(mockRepo)

	res, err := uc.GetSessionDetails(context.Background(), uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "invalid session ID", err.Error())
}

func TestGetSessionDetails_Success(t *testing.T) {
	mockRepo := new(MockSessionRepository)
	uc := usecase.NewSessionUseCase(mockRepo)
	sessionID := uuid.New()

	now := time.Now()
	expectedSession := &domain.InterviewSession{
		ID:        sessionID,
		Status:    domain.SessionStatusInProgress,
		StartedAt: &now,
	}

	mockRepo.On("GetByID", mock.Anything, sessionID).Return(expectedSession, nil)

	res, err := uc.GetSessionDetails(context.Background(), sessionID)

	assert.NoError(t, err)
	assert.Equal(t, expectedSession, res)
	mockRepo.AssertExpectations(t)
}
