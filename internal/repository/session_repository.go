package repository

import (
	"context"
	"encoding/json"
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

// GetDashboardMetrics retrieves the metrics for the HR dashboard.
func (r *pgSessionRepository) GetDashboardMetrics(ctx context.Context) (domain.DashboardMetrics, error) {
	var metrics domain.DashboardMetrics

	// Count total candidates
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM candidates`).Scan(&metrics.TotalCandidates)
	if err != nil {
		return metrics, fmt.Errorf("failed to count candidates: %w", err)
	}

	// Count active sessions
	err = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM interview_sessions WHERE status = $1`, domain.SessionStatusInProgress).Scan(&metrics.ActiveSessions)
	if err != nil {
		return metrics, fmt.Errorf("failed to count active sessions: %w", err)
	}

	// Count done sessions
	err = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM interview_sessions WHERE status = $1`, domain.SessionStatusDone).Scan(&metrics.DoneSessions)
	if err != nil {
		return metrics, fmt.Errorf("failed to count done sessions: %w", err)
	}

	return metrics, nil
}

// GetRecentInterviews retrieves a list of recent interviews.
func (r *pgSessionRepository) GetRecentInterviews(ctx context.Context, limit int) ([]domain.RecentInterview, error) {
	query := `
		SELECT s.id, c.name, j.title, s.status, s.score
		FROM interview_sessions s
		JOIN candidates c ON s.candidate_id = c.id
		JOIN jobs j ON s.job_id = j.id
		ORDER BY s.started_at DESC NULLS LAST
		LIMIT $1
	`
	rows, err := r.db.Pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent interviews: %w", err)
	}
	defer rows.Close()

	var interviews []domain.RecentInterview
	for rows.Next() {
		var interview domain.RecentInterview
		var score *string
		if err := rows.Scan(&interview.SessionID, &interview.CandidateName, &interview.JobTitle, &interview.Status, &score); err != nil {
			return nil, fmt.Errorf("failed to scan interview row: %w", err)
		}
		if score != nil {
			interview.Score = *score
		} else {
			interview.Score = "-"
		}
		interviews = append(interviews, interview)
	}

	return interviews, nil
}

// GetJobs retrieves all jobs.
func (r *pgSessionRepository) GetJobs(ctx context.Context) ([]domain.Job, error) {
	query := `SELECT id, title, department, interview_config, created_at FROM jobs ORDER BY created_at DESC`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %w", err)
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var job domain.Job
		var configBytes []byte
		if err := rows.Scan(&job.ID, &job.Title, &job.Department, &configBytes, &job.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		if len(configBytes) > 0 {
			_ = json.Unmarshal(configBytes, &job.InterviewConfig)
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// GetInterviewsByJobID retrieves all interviews for a specific job
func (r *pgSessionRepository) GetInterviewsByJobID(ctx context.Context, jobID uuid.UUID) ([]domain.RecentInterview, error) {
	query := `
		SELECT s.id, c.name, j.title, s.status, s.score
		FROM interview_sessions s
		JOIN candidates c ON s.candidate_id = c.id
		JOIN jobs j ON s.job_id = j.id
		WHERE s.job_id = $1
		ORDER BY s.started_at DESC NULLS LAST
	`
	rows, err := r.db.Pool.Query(ctx, query, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to query interviews by job: %w", err)
	}
	defer rows.Close()

	var interviews []domain.RecentInterview
	for rows.Next() {
		var interview domain.RecentInterview
		var score *string
		if err := rows.Scan(&interview.SessionID, &interview.CandidateName, &interview.JobTitle, &interview.Status, &score); err != nil {
			return nil, fmt.Errorf("failed to scan interview row: %w", err)
		}
		if score != nil {
			interview.Score = *score
		} else {
			interview.Score = "-"
		}
		interviews = append(interviews, interview)
	}

	return interviews, nil
}

// GetJobByID retrieves a specific job.
func (r *pgSessionRepository) GetJobByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	query := `SELECT id, title, department, interview_config, created_at FROM jobs WHERE id = $1`
	var job domain.Job
	var configBytes []byte
	
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&job.ID, &job.Title, &job.Department, &configBytes, &job.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to query job by id: %w", err)
	}
	
	if len(configBytes) > 0 {
		_ = json.Unmarshal(configBytes, &job.InterviewConfig)
	}
	return &job, nil
}

// CreateJob creates a new job.
func (r *pgSessionRepository) CreateJob(ctx context.Context, job *domain.Job) error {
	configBytes, _ := json.Marshal(job.InterviewConfig)
	query := `INSERT INTO jobs (id, title, department, interview_config, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Pool.Exec(ctx, query, job.ID, job.Title, job.Department, configBytes, job.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert job: %w", err)
	}
	return nil
}

// GetSessionReport retrieves the complete interview report for the HR view.
func (r *pgSessionRepository) GetSessionReport(ctx context.Context, sessionID uuid.UUID) (domain.ReportData, error) {
	var report domain.ReportData

	// 1. Get Session details
	session, err := r.GetByID(ctx, sessionID)
	if err != nil {
		return report, fmt.Errorf("failed to get session for report: %w", err)
	}
	report.Session = *session

	// 2. Get Candidate details
	candidateQuery := `SELECT id, name, email, job_id, created_at FROM candidates WHERE id = $1`
	err = r.db.Pool.QueryRow(ctx, candidateQuery, session.CandidateID).Scan(
		&report.Candidate.ID,
		&report.Candidate.Name,
		&report.Candidate.Email,
		&report.Candidate.JobID,
		&report.Candidate.ExpiresAt, // Using ExpiresAt to hold created_at for simplicity in this view
	)
	if err != nil {
		return report, fmt.Errorf("failed to get candidate for report: %w", err)
	}

	// 3. Get Job details
	jobQuery := `SELECT id, title, department, interview_config FROM jobs WHERE id = $1`
	var configBytes []byte
	err = r.db.Pool.QueryRow(ctx, jobQuery, session.JobID).Scan(
		&report.Job.ID,
		&report.Job.Title,
		&report.Job.Department,
		&configBytes,
	)
	if err != nil {
		return report, fmt.Errorf("failed to get job for report: %w", err)
	}
	if len(configBytes) > 0 {
		_ = json.Unmarshal(configBytes, &report.Job.InterviewConfig)
	}

	// 4. Get Turns (Transcript)
	turnsQuery := `SELECT id, role, content, turn_index, audio_offset_ms, duration_ms FROM session_turns WHERE session_id = $1 ORDER BY turn_index ASC`
	rows, err := r.db.Pool.Query(ctx, turnsQuery, sessionID)
	if err != nil {
		return report, fmt.Errorf("failed to query turns for report: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var turn domain.SessionTurn
		turn.SessionID = sessionID
		if err := rows.Scan(&turn.ID, &turn.Role, &turn.Content, &turn.TurnIndex, &turn.AudioOffsetMs, &turn.DurationMs); err != nil {
			return report, fmt.Errorf("failed to scan turn: %w", err)
		}
		report.Turns = append(report.Turns, turn)
	}

	return report, nil
}

// CreateCandidate creates a new candidate in the database.
func (r *pgSessionRepository) CreateCandidate(ctx context.Context, candidate *domain.Candidate) error {
	query := `
		INSERT INTO candidates (id, name, email, job_id, created_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		candidate.ID,
		candidate.Name,
		candidate.Email,
		candidate.JobID,
	)

	if err != nil {
		logger.Error("Failed to create candidate", zap.Error(err), zap.String("candidate_id", candidate.ID.String()))
		return fmt.Errorf("repository: failed to create candidate: %w", err)
	}

	return nil
}
