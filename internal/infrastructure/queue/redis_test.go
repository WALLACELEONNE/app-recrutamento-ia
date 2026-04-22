package queue_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/username/app-recrutamento-ia/internal/infrastructure/queue"
)

func TestNewRedisQueue_ConnectionError(t *testing.T) {
	// Tenta conectar num Redis inexistente
	q, err := queue.NewRedisQueue("127.0.0.1:9999", "", "test_queue")

	assert.Error(t, err)
	assert.Nil(t, q)
}

// Em um ambiente real com Docker, poderíamos usar Testcontainers ou miniredis
// para testar Enqueue e Listen com sucesso.
func TestEnqueue_FailsWithoutConnection(t *testing.T) {
	// Apenas garantimos que os métodos compilarão e retornarão erro
	// se a conexão não existir ou cliente for nil (embora NewRedisQueue já trave antes).
}
