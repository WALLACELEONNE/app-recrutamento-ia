package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/username/app-recrutamento-ia/internal/domain"
)

// SessionUseCase defines the interface for interview session business logic
type SessionUseCase interface {
	StartSession(ctx context.Context, sessionID uuid.UUID) error
	GetSessionDetails(ctx context.Context, sessionID uuid.UUID) (*domain.InterviewSession, error)
}

type sessionUseCase struct {
	repo domain.SessionRepository
}

// NewSessionUseCase creates a new instance of SessionUseCase
func NewSessionUseCase(repo domain.SessionRepository) SessionUseCase {
	return &sessionUseCase{
		repo: repo,
	}
}

// StartSession marks a session as "in_progress" and sets the start time
func (uc *sessionUseCase) StartSession(ctx context.Context, sessionID uuid.UUID) error {
	if sessionID == uuid.Nil {
		return errors.New("invalid session ID")
	}

	session, err := uc.repo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.Status == domain.SessionStatusDone {
		return errors.New("session is already done")
	}

	// Em um cenário real de banco, isso idealmente faria um Update direto ou Transaction
	err = uc.repo.UpdateStatus(ctx, sessionID, domain.SessionStatusInProgress)
	if err != nil {
		return err
	}

	return nil
}

// GetSessionDetails retrieves the interview session details
func (uc *sessionUseCase) GetSessionDetails(ctx context.Context, sessionID uuid.UUID) (*domain.InterviewSession, error) {
	if sessionID == uuid.Nil {
		return nil, errors.New("invalid session ID")
	}

	return uc.repo.GetByID(ctx, sessionID)
}
