package openai_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/username/app-recrutamento-ia/internal/domain"
	"github.com/username/app-recrutamento-ia/internal/infrastructure/openai"
)

func TestNewLLMClient(t *testing.T) {
	// Falha sem API key
	client, err := openai.NewLLMClient("")
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, "openai API key is required", err.Error())

	// Sucesso com API key
	client, err = openai.NewLLMClient("dummy-key")
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestGenerateResponseStream_EmptyPrompt(t *testing.T) {
	client, _ := openai.NewLLMClient("dummy-key")

	stream, err := client.GenerateResponseStream(context.Background(), "", []domain.SessionTurn{}, "Olá")
	assert.Error(t, err)
	assert.Nil(t, stream)
	assert.Equal(t, "system prompt cannot be empty", err.Error())
}
