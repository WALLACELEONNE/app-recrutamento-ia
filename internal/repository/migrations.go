package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/username/app-recrutamento-ia/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Migrate creates the necessary tables if they do not exist
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS jobs (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			title VARCHAR(255) NOT NULL,
			department VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS candidates (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			job_id UUID REFERENCES jobs(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS interview_sessions (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			candidate_id UUID REFERENCES candidates(id),
			job_id UUID REFERENCES jobs(id),
			status VARCHAR(50) NOT NULL,
			score VARCHAR(50),
			started_at TIMESTAMP WITH TIME ZONE,
			ended_at TIMESTAMP WITH TIME ZONE,
			duration_s INT,
			audio_s3_key TEXT,
			transcript_s3_key TEXT
		);`,
		`ALTER TABLE jobs ADD COLUMN IF NOT EXISTS interview_config JSONB DEFAULT '{}'::jsonb;`,
		`CREATE TABLE IF NOT EXISTS session_turns (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			session_id UUID REFERENCES interview_sessions(id),
			role VARCHAR(50) NOT NULL,
			content TEXT NOT NULL,
			turn_index INT NOT NULL,
			audio_offset_ms INT,
			duration_ms INT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, query := range queries {
		if _, err := pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	logger.Info("Database migrations completed successfully")
	return nil
}

// SeedAdmin creates a default admin user if no users exist in the database
func SeedAdmin(ctx context.Context, pool *pgxpool.Pool) error {
	var count int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", "admin@acme.com").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check users count: %w", err)
	}

	if count > 0 {
		logger.Info("Admin user already exists, skipping admin seeding")
		return nil
	}

	// Create default admin
	adminEmail := "admin@acme.com"
	adminPassword := "password123"

	// Hash password
	bytes, err := bcrypt.GenerateFromPassword([]byte(adminPassword), 10)
	if err != nil {
		return fmt.Errorf("failed to hash default admin password: %w", err)
	}

	id := uuid.New()
	now := time.Now()

	_, err = pool.Exec(ctx,
		`INSERT INTO users (id, name, email, password_hash, role, created_at, updated_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, "Admin Default", adminEmail, string(bytes), "admin", now, now,
	)

	if err != nil {
		return fmt.Errorf("failed to seed admin user: %w", err)
	}

	logger.Info("Default admin user seeded successfully", zap.String("email", adminEmail))
	return nil
}

// SeedDemoData populates the database with some demo jobs, candidates, and interview sessions
// This allows testing the UI without having to create everything from scratch
func SeedDemoData(ctx context.Context, pool *pgxpool.Pool) error {
	var jobCount int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM jobs").Scan(&jobCount)
	if err != nil {
		return fmt.Errorf("failed to check jobs count: %w", err)
	}

	// Only seed if the database is empty
	if jobCount > 0 {
		logger.Info("Demo data already exists, skipping demo seeding")
		return nil
	}

	logger.Info("Seeding demo data (Jobs, Candidates, and Sessions)...")

	// 1. Create Jobs
	job1ID := uuid.New()
	job2ID := uuid.New()
	now := time.Now()

	job1Config := `{"n_questions": 5, "max_minutes": 15, "persona": "Tech Recruiter Senior", "tone": "profissional", "focus_areas": ["Golang", "Microservices", "AWS"]}`
	job2Config := `{"n_questions": 3, "max_minutes": 10, "persona": "Data Lead", "tone": "tecnico", "focus_areas": ["Python", "SQL", "Airflow"]}`

	_, err = pool.Exec(ctx, `
		INSERT INTO jobs (id, title, department, interview_config, created_at)
		VALUES 
		($1, 'Desenvolvedor Golang Sênior', 'Engenharia', $3, $5),
		($2, 'Engenheiro de Dados Pleno', 'Dados', $4, $5)
	`, job1ID, job2ID, job1Config, job2Config, now)

	if err != nil {
		return fmt.Errorf("failed to seed demo jobs: %w", err)
	}

	// 2. Create Candidates
	cand1ID := uuid.New()
	cand2ID := uuid.New()
	cand3ID := uuid.New()

	_, err = pool.Exec(ctx, `
		INSERT INTO candidates (id, name, email, job_id, created_at)
		VALUES 
		($1, 'Carlos Eduardo Silva', 'carlos.dev@example.com', $4, $7),
		($2, 'Mariana Costa', 'mariana.data@example.com', $5, $7),
		($3, 'João Pedro Santos', 'joao.go@example.com', $4, $7)
	`, cand1ID, cand2ID, cand3ID, job1ID, job2ID, now)

	if err != nil {
		return fmt.Errorf("failed to seed demo candidates: %w", err)
	}

	// 3. Create Sessions
	sess1ID := uuid.New()
	sess2ID := uuid.New()
	sess3ID := uuid.New()

	startedSess1 := now.Add(-2 * time.Hour)
	endedSess1 := startedSess1.Add(15 * time.Minute)

	startedSess2 := now.Add(-30 * time.Minute)

	_, err = pool.Exec(ctx, `
		INSERT INTO interview_sessions (id, candidate_id, job_id, status, score, started_at, ended_at, duration_s)
		VALUES 
		($1, $4, $7, 'done', '8.5/10', $9, $10, 900),
		($2, $5, $8, 'in_progress', null, $11, null, null),
		($3, $6, $7, 'invited', null, null, null, null)
	`, sess1ID, sess2ID, sess3ID, cand1ID, cand2ID, cand3ID, job1ID, job2ID, startedSess1, endedSess1, startedSess2)

	if err != nil {
		return fmt.Errorf("failed to seed demo sessions: %w", err)
	}

	// 4. Create Session Turns (Transcript) for the completed session
	_, err = pool.Exec(ctx, `
		INSERT INTO session_turns (session_id, role, content, turn_index, audio_offset_ms, duration_ms)
		VALUES 
		($1, 'ai', 'Olá Carlos, seja bem-vindo. Sou a Nova Voice AI e conduzirei sua entrevista para a vaga de Desenvolvedor Golang. Podemos começar?', 1, 0, 5000),
		($1, 'candidate', 'Olá! Sim, podemos começar.', 2, 5500, 2000),
		($1, 'ai', 'Ótimo! Me conte um pouco sobre sua experiência arquitetando microsserviços em Go e como você lida com comunicação assíncrona.', 3, 8000, 7000),
		($1, 'candidate', 'Eu tenho 4 anos de experiência com Go. Costumo usar RabbitMQ ou Kafka para mensageria, e gRPC para comunicação síncrona entre serviços mais críticos.', 4, 16000, 12000),
		($1, 'ai', 'Excelente. E como você lida com a idempotência no consumo dessas mensagens?', 5, 29000, 4000)
	`, sess1ID)

	if err != nil {
		return fmt.Errorf("failed to seed demo turns: %w", err)
	}

	logger.Info("Demo data seeded successfully")
	return nil
}
