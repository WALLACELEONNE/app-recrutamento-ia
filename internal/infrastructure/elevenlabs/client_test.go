package elevenlabs_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/username/app-recrutamento-ia/internal/infrastructure/elevenlabs"
)

func TestNewTTSClient_ErrorEmptyKey(t *testing.T) {
	client, err := elevenlabs.NewTTSClient("", "voice-id")
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, "elevenlabs API key is required", err.Error())
}

func TestNewTTSClient_ErrorEmptyVoiceID(t *testing.T) {
	client, err := elevenlabs.NewTTSClient("test-key", "")
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, "elevenlabs voice ID is required", err.Error())
}

func TestNewTTSClient_Success(t *testing.T) {
	client, err := elevenlabs.NewTTSClient("test-key", "test-voice-id")
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestSynthesizeStream_Timeout(t *testing.T) {
	// A chave e ID são falsos, a chamada HTTP falhará, mas o foco é no comportamento do stream
	client, _ := elevenlabs.NewTTSClient("invalid-key", "invalid-voice")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	textStream := make(chan string)
	
	// Fechamos o stream simulando que nada mais virá. O cliente tentará
	// chamar a API para enviar as sobras e falhará silenciosamente (apenas log),
	// mas não causará panic.
	close(textStream)

	audioStream, err := client.SynthesizeStream(ctx, textStream)
	assert.NoError(t, err)
	assert.NotNil(t, audioStream)

	// Lê tudo do audioStream para evitar deadlock
	for range audioStream {
	}
}

// Para testar a heurística de pontuação (função isPunctuation que não está exportada,
// testamos indiretamente através de injeção de tokens no stream).
func TestSynthesizeStream_PunctuationTrigger(t *testing.T) {
	client, _ := elevenlabs.NewTTSClient("invalid-key", "invalid-voice")
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	textStream := make(chan string)
	audioStream, err := client.SynthesizeStream(ctx, textStream)
	assert.NoError(t, err)

	// Simula tokens vindo do LLM
	textStream <- "Olá, "
	textStream <- "como você "
	textStream <- "está?" // Pontuação deve disparar a síntese (com erro no backend, mas o fluxo segue)

	// Dá um tempo curto para goroutine rodar
	time.Sleep(100 * time.Millisecond)

	// Se houvesse sucesso na API, audioStream receberia os bytes.
	// Como a key é inválida, o método 'synthesizeAndSend' apenas fará um log de erro.
	// Aqui verificamos apenas se não houve deadlock.
	close(textStream)

	for range audioStream {
	}
}
