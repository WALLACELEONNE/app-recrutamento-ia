package deepgram_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/username/app-recrutamento-ia/internal/infrastructure/deepgram"
)

func TestNewSTTClient(t *testing.T) {
	// Falha sem API key
	client, err := deepgram.NewSTTClient("")
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, "deepgram API key is required", err.Error())

	// Sucesso com API key
	client, err = deepgram.NewSTTClient("dummy-key")
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
