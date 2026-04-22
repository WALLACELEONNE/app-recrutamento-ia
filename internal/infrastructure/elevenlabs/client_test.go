package elevenlabs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/username/app-recrutamento-ia/internal/infrastructure/elevenlabs"
)

func TestNewTTSClient(t *testing.T) {
	// Falha sem API key
	client, err := elevenlabs.NewTTSClient("", "voice-id")
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, "elevenlabs API key is required", err.Error())

	// Falha sem Voice ID
	client, err = elevenlabs.NewTTSClient("api-key", "")
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, "elevenlabs voice ID is required", err.Error())

	// Sucesso com todos os params
	client, err = elevenlabs.NewTTSClient("dummy-key", "voice-id")
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
