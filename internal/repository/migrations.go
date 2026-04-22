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
		// We can add other tables here later (candidates, interviews)
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
