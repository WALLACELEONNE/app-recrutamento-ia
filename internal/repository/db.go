package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/username/app-recrutamento-ia/internal/logger"
)

// DB representa a conexão com o banco de dados via pgxpool
type DB struct {
	Pool *pgxpool.Pool
}

// NewDB cria uma nova conexão com o PostgreSQL
func NewDB(ctx context.Context, dsn string) (*DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("database DSN is empty")
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configurações de pool otimizadas
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Ping para garantir que a conexão está viva
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to PostgreSQL successfully")

	return &DB{Pool: pool}, nil
}

// Close encerra a conexão com o banco
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		logger.Info("Database connection pool closed")
	}
}
