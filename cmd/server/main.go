package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/username/app-recrutamento-ia/internal/handlers"
	"github.com/username/app-recrutamento-ia/internal/logger"
	"github.com/username/app-recrutamento-ia/internal/repository"
	"github.com/username/app-recrutamento-ia/internal/server"
	"github.com/username/app-recrutamento-ia/internal/usecase"
	"go.uber.org/zap"
)

// @title App Recrutamento IA - API
// @version 1.0
// @description Plataforma SaaS de entrevistas por voz com IA
// @host 127.0.0.1:3000
// @BasePath /api/v1
func main() {
	_ = godotenv.Load()

	logger.InitLogger()
	defer logger.Get().Sync()

	port := getEnvOrDefault("PORT", "3000")

	// Dependências da Clean Architecture
	dbURL := getEnvOrDefault("DATABASE_URL", "postgres://admin:password@127.0.0.1:5432/recrutamento_db?sslmode=disable")
	db, err := repository.NewDB(context.Background(), dbURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Execute Automigrations and Seed
	if err := repository.Migrate(context.Background(), db.Pool); err != nil {
		logger.Fatal("Database migration failed", zap.Error(err))
	}
	if err := repository.SeedAdmin(context.Background(), db.Pool); err != nil {
		logger.Fatal("Admin seeding failed", zap.Error(err))
	}

	sessionRepo := repository.NewSessionRepository(db)
	sessionUC := usecase.NewSessionUseCase(sessionRepo)
	sessionHandler := handlers.NewSessionHandler(sessionUC)
	frontendHandler := handlers.NewFrontendHandler()

	// Initialize Auth Module
	userRepo := repository.NewUserRepository(db.Pool)
	authUC := usecase.NewAuthUseCase(userRepo)
	authHandler := handlers.NewAuthHandler(authUC)

	// Inicializa o roteador com os handlers injetados
	router := server.NewRouter(sessionHandler, frontendHandler, authHandler)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
		// Configurações de timeout para segurança e prevenção contra Slowloris
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Canal para interceptar sinais do SO (Graceful Shutdown)
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("Starting HTTP server", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Could not listen on port", zap.Error(err))
		}
	}()

	<-stopChan
	logger.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited properly")
}

func getEnvOrDefault(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
