package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/username/app-recrutamento-ia/internal/infrastructure/openai"
	"github.com/username/app-recrutamento-ia/internal/infrastructure/queue"
	zlog "github.com/username/app-recrutamento-ia/internal/logger"
	"github.com/username/app-recrutamento-ia/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type AnalysisJob struct {
	SessionID string `json:"session_id"`
}

// rateLimiter middleware básico (60 req/minuto)
var limiter = rate.NewLimiter(1, 60)

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// @title Analysis Worker API
// @version 1.0
// @description API independente para o Worker de Análise (LLM). Fornece endpoints de métricas e healthcheck.
func main() {
	_ = godotenv.Load()
	zlog.InitLogger()
	defer zlog.Get().Sync()

	port := getEnvOrDefault("WORKER_PORT", "4000")
	dbURL := getEnvOrDefault("DATABASE_URL", "postgres://admin:password@127.0.0.1:5432/recrutamento_db?sslmode=disable")
	redisAddr := getEnvOrDefault("REDIS_ADDR", "127.0.0.1:6379")

	// Dependências (Database)
	db, err := repository.NewDB(context.Background(), dbURL)
	if err != nil {
		zlog.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// OpenAI Client
	_, err = openai.NewLLMClient(getEnvOrDefault("OPENAI_API_KEY", "dummy"))
	if err != nil {
		zlog.Fatal("Failed to init LLM")
	}

	// Queue (Redis)
	redisQueue, err := queue.NewRedisQueue(redisAddr, "", "analysis_jobs")
	if err != nil {
		zlog.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisQueue.Close()

	// Processador em Background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go redisQueue.Listen(ctx, func(ctx context.Context, jobData []byte) error {
		var job AnalysisJob
		if err := json.Unmarshal(jobData, &job); err != nil {
			return err
		}

		zlog.Info("Processando análise de sessão", zap.String("session_id", job.SessionID))
		
		// Simulação de tempo de processamento assíncrono (LLM prompt etc)
		time.Sleep(2 * time.Second)
		
		// O processo de análise real envolveria:
		// 1. Buscar `session_turns` no Postgres
		// 2. Montar texto completo da entrevista
		// 3. Enviar prompt para LLMClient
		// 4. Receber score, skills (JSONB)
		// 5. Atualizar tabela `interview_sessions` com Status='done' e Score final.
		
		zlog.Info("Análise concluída com sucesso", zap.String("session_id", job.SessionID))
		return nil
	})

	// API HTTP RESTful independente
	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Recoverer)
	router.Use(RateLimitMiddleware)

	// Endpoints Base
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Worker OK"))
	})
	
	// Observabilidade
	router.Get("/metrics", promhttp.Handler().ServeHTTP)

	// Endpoint para enfileirar manualmente uma análise (útil para retries forçados)
	router.Post("/api/v1/analyze", func(w http.ResponseWriter, r *http.Request) {
		var payload AnalysisJob
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
		
		if payload.SessionID == "" {
			http.Error(w, "session_id is required", http.StatusBadRequest)
			return
		}

		err := redisQueue.Enqueue(r.Context(), payload)
		if err != nil {
			http.Error(w, "Failed to enqueue job", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"enqueued"}`))
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful Shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		zlog.Info("Analysis Worker API HTTP server running", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zlog.Fatal("Could not listen on port", zap.Error(err))
		}
	}()

	<-stopChan
	zlog.Info("Shutting down Analysis Worker gracefully...")

	// Timeout de 15 segundos para encerrar os jobs em processamento e HTTP
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zlog.Fatal("Worker forced to shutdown", zap.Error(err))
	}

	zlog.Info("Analysis Worker exited properly")
}

func getEnvOrDefault(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
