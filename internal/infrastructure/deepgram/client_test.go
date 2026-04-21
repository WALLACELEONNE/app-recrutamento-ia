package deepgram

import (
	"context"
	"testing"
	"time"

	apiinterfaces "github.com/deepgram/deepgram-go-sdk/pkg/api/listen/v1/websocket/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestNewSTTClient_Error(t *testing.T) {
	client, err := NewSTTClient("")
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, "deepgram API key is required", err.Error())
}

func TestNewSTTClient_Success(t *testing.T) {
	client, err := NewSTTClient("test-api-key")
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestTranscribeStream_Timeout(t *testing.T) {
	client, _ := NewSTTClient("invalid-key")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	audioStream := make(chan []byte)
	_, err := client.TranscribeStream(ctx, audioStream)

	assert.Error(t, err)
}

func TestDgCallback_Message(t *testing.T) {
	outStream := make(chan string, 5)
	cb := &dgCallback{outStream: outStream}

	// Teste com mensagem válida final
	mr := &apiinterfaces.MessageResponse{
		IsFinal: true,
		Channel: apiinterfaces.Channel{
			Alternatives: []apiinterfaces.Alternative{
				{Transcript: "hello world"},
			},
		},
	}

	err := cb.Message(mr)
	assert.NoError(t, err)

	select {
	case msg := <-outStream:
		assert.Equal(t, "hello world", msg)
	default:
		t.Fatal("Expected message in channel")
	}

	// Teste com mensagem não final (não deve ir pro canal)
	mrNotFinal := &apiinterfaces.MessageResponse{
		IsFinal: false,
		Channel: apiinterfaces.Channel{
			Alternatives: []apiinterfaces.Alternative{
				{Transcript: "hello"},
			},
		},
	}
	err = cb.Message(mrNotFinal)
	assert.NoError(t, err)

	select {
	case <-outStream:
		t.Fatal("Did not expect message in channel")
	default:
	}

	// Test empty message
	mrEmpty := &apiinterfaces.MessageResponse{
		IsFinal: true,
		Channel: apiinterfaces.Channel{
			Alternatives: []apiinterfaces.Alternative{
				{Transcript: ""},
			},
		},
	}
	err = cb.Message(mrEmpty)
	assert.NoError(t, err)

	select {
	case <-outStream:
		t.Fatal("Did not expect empty message in channel")
	default:
	}
}

func TestDgCallback_OtherMethods(t *testing.T) {
	cb := &dgCallback{}
	assert.NoError(t, cb.Open(nil))
	assert.NoError(t, cb.Metadata(nil))
	assert.NoError(t, cb.SpeechStarted(nil))
	assert.NoError(t, cb.UtteranceEnd(nil))
	assert.NoError(t, cb.Close(nil))
	assert.NoError(t, cb.UnhandledEvent(nil))

	errResp := &apiinterfaces.ErrorResponse{ErrMsg: "test error"}
	assert.NoError(t, cb.Error(errResp))
}
