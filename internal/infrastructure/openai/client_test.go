package openai_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/username/app-recrutamento-ia/internal/domain"
	oainfra "github.com/username/app-recrutamento-ia/internal/infrastructure/openai"
)

func TestNewLLMClient_Error(t *testing.T) {
	client, err := oainfra.NewLLMClient("")
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, "openai API key is required", err.Error())
}

func TestNewLLMClient_Success(t *testing.T) {
	client, err := oainfra.NewLLMClient("test-api-key")
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestGenerateResponseStream_EmptySystemPrompt(t *testing.T) {
	client, _ := oainfra.NewLLMClient("test-key")

	ctx := context.Background()
	_, err := client.GenerateResponseStream(ctx, "", nil, "hello")

	assert.Error(t, err)
	assert.Equal(t, "system prompt cannot be empty", err.Error())
}

func TestGenerateResponseStream_InvalidKey(t *testing.T) {
	client, _ := oainfra.NewLLMClient("invalid-key")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	history := []domain.SessionTurn{
		{Role: domain.RoleAI, Content: "Hello, how are you?"},
	}

	_, err := client.GenerateResponseStream(ctx, "You are an HR bot", history, "I am fine")

	// Since the key is invalid, the HTTP request should fail or return 401
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create completion stream")
}
