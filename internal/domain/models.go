package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SessionStatus represents the current state of an interview session.
type SessionStatus string

const (
	SessionStatusInvited    SessionStatus = "invited"
	SessionStatusInProgress SessionStatus = "in_progress"
	SessionStatusDone       SessionStatus = "done"
)

// Role represents who is speaking in a session turn.
type Role string

const (
	RoleAI        Role = "ai"
	RoleCandidate Role = "candidate"
)

// Organization represents a multi-tenant client of the platform.
type Organization struct {
	ID         uuid.UUID `json:"id"`
	Slug       string    `json:"slug"`
	Plan       string    `json:"plan"`
	SchemaName string    `json:"schema_name"`
	CreatedAt  time.Time `json:"created_at"`
}

// User represents a system administrator or HR recruiter.
type User struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"`
	Role           string    `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
}

// Job represents an open position with interview configurations.
type Job struct {
	ID              uuid.UUID `json:"id"`
	Title           string    `json:"title"`
	Department      string    `json:"department"`
	InterviewConfig JobConfig `json:"interview_config"`
	CreatedAt       time.Time `json:"created_at"`
}

// JobConfig defines the AI parameters for the interview.
type JobConfig struct {
	NQuestions int      `json:"n_questions"`
	MaxMinutes int      `json:"max_minutes"`
	Persona    string   `json:"persona"`
	Tone       string   `json:"tone"`
	FocusAreas []string `json:"focus_areas"`
}

// Candidate represents a person applying for a job.
type Candidate struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Phone         string     `json:"phone"`
	JobID         uuid.UUID  `json:"job_id"`
	InviteToken   uuid.UUID  `json:"invite_token"`
	ExpiresAt     time.Time  `json:"expires_at"`
	GDPRConsentAt *time.Time `json:"gdpr_consent_at,omitempty"`
	AnonymizedAt  *time.Time `json:"anonymized_at,omitempty"`
}

// InterviewSession represents a live or finished interview.
type InterviewSession struct {
	ID              uuid.UUID     `json:"id"`
	CandidateID     uuid.UUID     `json:"candidate_id"`
	JobID           uuid.UUID     `json:"job_id"`
	Status          SessionStatus `json:"status"`
	StartedAt       *time.Time    `json:"started_at,omitempty"`
	EndedAt         *time.Time    `json:"ended_at,omitempty"`
	DurationSeconds int           `json:"duration_s,omitempty"`
	AudioS3Key      string        `json:"audio_s3_key,omitempty"`
	TranscriptS3Key string        `json:"transcript_s3_key,omitempty"`
}

// SessionTurn is a single interaction (speech) during the interview.
type SessionTurn struct {
	ID            uuid.UUID `json:"id"`
	SessionID     uuid.UUID `json:"session_id"`
	Role          Role      `json:"role"`
	Content       string    `json:"content"`
	TurnIndex     int       `json:"turn_index"`
	AudioOffsetMs int       `json:"audio_offset_ms"`
	DurationMs    int       `json:"duration_ms"`
}

// DashboardMetrics holds the aggregated data for the HR dashboard
type DashboardMetrics struct {
	TotalCandidates int
	ActiveSessions  int
	DoneSessions    int
}

// RecentInterview holds the data for the recent interviews table
type RecentInterview struct {
	CandidateName string
	JobTitle      string
	Status        SessionStatus
	Score         string
}

// DashboardData holds the view model for the dashboard page
type DashboardData struct {
	Metrics          DashboardMetrics
	RecentInterviews []RecentInterview
}

// --- Interfaces for Clean Architecture ---

// SessionRepository defines the database operations for sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *InterviewSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*InterviewSession, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status SessionStatus) error
	GetDashboardMetrics(ctx context.Context) (DashboardMetrics, error)
	GetRecentInterviews(ctx context.Context, limit int) ([]RecentInterview, error)
	GetJobs(ctx context.Context) ([]Job, error)
	CreateJob(ctx context.Context, job *Job) error
}

// TurnRepository defines the database operations for conversation turns.
type TurnRepository interface {
	AddTurn(ctx context.Context, turn *SessionTurn) error
	GetTurnsBySession(ctx context.Context, sessionID uuid.UUID) ([]*SessionTurn, error)
}

// UserRepository defines the database operations for users.
type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

// STTClient is the interface for Speech-to-Text providers (e.g., Deepgram).
type STTClient interface {
	TranscribeStream(ctx context.Context, audioStream <-chan []byte) (<-chan string, error)
}

// LLMClient is the interface for Language Models (e.g., OpenAI).
type LLMClient interface {
	GenerateResponseStream(ctx context.Context, systemPrompt string, history []SessionTurn, currentTurn string) (<-chan string, error)
}

// TTSClient is the interface for Text-to-Speech providers (e.g., ElevenLabs).
type TTSClient interface {
	SynthesizeStream(ctx context.Context, textStream <-chan string) (<-chan []byte, error)
}
