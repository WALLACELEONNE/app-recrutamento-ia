package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/username/app-recrutamento-ia/internal/logger"
	"go.uber.org/zap"
)

// RedisQueue implementa uma fila de processamento básica usando Redis Lists.
type RedisQueue struct {
	client *redis.Client
	queue  string
}

// NewRedisQueue inicializa a conexão com o Redis e retorna a fila.
func NewRedisQueue(addr, password, queueName string) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisQueue{
		client: client,
		queue:  queueName,
	}, nil
}

// Enqueue adiciona um payload (JSON serializado) ao final da fila.
func (rq *RedisQueue) Enqueue(ctx context.Context, payload interface{}) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = rq.client.RPush(ctx, rq.queue, bytes).Err()
	if err != nil {
		logger.Error("Failed to enqueue job", zap.Error(err), zap.String("queue", rq.queue))
		return err
	}

	return nil
}

// Listen processa a fila de forma contínua e assíncrona.
func (rq *RedisQueue) Listen(ctx context.Context, handler func(context.Context, []byte) error) {
	logger.Info("Started listening to queue", zap.String("queue", rq.queue))

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping queue listener", zap.String("queue", rq.queue))
			return
		default:
			// BLPop bloqueia até que um item esteja disponível ou dê timeout (2 segundos)
			result, err := rq.client.BLPop(ctx, 2*time.Second, rq.queue).Result()

			if err == redis.Nil {
				// Timeout normal, fila vazia
				continue
			} else if err != nil {
				// Ignora erro de context canceled no shutdown
				if ctx.Err() == nil {
					logger.Error("Error popping from queue", zap.Error(err), zap.String("queue", rq.queue))
					time.Sleep(1 * time.Second) // Backoff on error
				}
				continue
			}

			if len(result) == 2 {
				jobData := []byte(result[1])
				
				// Processa o job de forma bloqueante neste worker (pode ser go handler() para concorrencia)
				err := handler(ctx, jobData)
				if err != nil {
					logger.Error("Job processing failed", zap.Error(err))
					// Em um sistema real como RabbitMQ, faríamos NACK. 
					// Aqui poderíamos enviar para uma Dead Letter Queue (DLQ).
				}
			}
		}
	}
}

// Close encerra a conexão com o Redis.
func (rq *RedisQueue) Close() error {
	return rq.client.Close()
}
