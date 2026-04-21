package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/username/app-recrutamento-ia/internal/domain"
	"github.com/username/app-recrutamento-ia/internal/logger"
	"go.uber.org/zap"
)

// pgSessionRepository implementa a interface domain.SessionRepository
type pgSessionRepository struct {
	db *DB
}

// NewSessionRepository cria uma nova instância do repositório de sessões
func NewSessionRepository(db *DB) domain.SessionRepository {
	return &pgSessionRepository{db: db}
}

// Create insere uma nova sessão de entrevista no banco de dados.
func (r *pgSessionRepository) Create(ctx context.Context, session *domain.InterviewSession) error {
	query := `
		INSERT INTO interview_sessions (id, candidate_id, job_id, status, started_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		session.ID,
		session.CandidateID,
		session.JobID,
		session.Status,
		session.StartedAt,
	)

	if err != nil {
		logger.Error("Failed to create session", zap.Error(err), zap.String("session_id", session.ID.String()))
		return fmt.Errorf("repository: failed to create session: %w", err)
	}

	return nil
}

// GetByID busca uma sessão pelo seu UUID.
func (r *pgSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.InterviewSession, error) {
	query := `
		SELECT id, candidate_id, job_id, status, started_at, ended_at, duration_s, audio_s3_key, transcript_s3_key
		FROM interview_sessions
		WHERE id = $1
	`
	row := r.db.Pool.QueryRow(ctx, query, id)

	var s domain.InterviewSession
	err := row.Scan(
		&s.ID,
		&s.CandidateID,
		&s.JobID,
		&s.Status,
		&s.StartedAt,
		&s.EndedAt,
		&s.DurationSeconds,
		&s.AudioS3Key,
		&s.TranscriptS3Key,
	)

	if err != nil {
		return nil, fmt.Errorf("repository: failed to get session: %w", err)
	}

	return &s, nil
}

// UpdateStatus atualiza o status de uma sessão existente.
func (r *pgSessionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SessionStatus) error {
	query := `
		UPDATE interview_sessions
		SET status = $1
		WHERE id = $2
	`
	tag, err := r.db.Pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("repository: failed to update session status: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("repository: session not found for update")
	}

	return nil
}
