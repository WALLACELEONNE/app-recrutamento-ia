package deepgram

import (
	"context"
	"errors"
	"fmt"

	apiinterfaces "github.com/deepgram/deepgram-go-sdk/pkg/api/listen/v1/websocket/interfaces"
	"github.com/deepgram/deepgram-go-sdk/pkg/client/interfaces"
	listen "github.com/deepgram/deepgram-go-sdk/pkg/client/listen/v1/websocket"
	"github.com/username/app-recrutamento-ia/internal/logger"
	"go.uber.org/zap"
)

// STTClient implementa a interface domain.STTClient para o Deepgram.
type STTClient struct {
	apiKey string
}

// NewSTTClient cria uma nova instância do cliente Deepgram.
func NewSTTClient(apiKey string) (*STTClient, error) {
	if apiKey == "" {
		return nil, errors.New("deepgram API key is required")
	}

	return &STTClient{
		apiKey: apiKey,
	}, nil
}

// dgCallback implementa a interface apiinterfaces.LiveMessageCallback do Deepgram SDK.
type dgCallback struct {
	outStream chan string
}

func (c *dgCallback) Open(or *apiinterfaces.OpenResponse) error {
	logger.Debug("Deepgram connection opened")
	return nil
}

func (c *dgCallback) Message(mr *apiinterfaces.MessageResponse) error {
	if mr != nil && len(mr.Channel.Alternatives) > 0 {
		transcript := mr.Channel.Alternatives[0].Transcript
		if transcript != "" && mr.IsFinal {
			c.outStream <- transcript
		}
	}
	return nil
}

func (c *dgCallback) Metadata(md *apiinterfaces.MetadataResponse) error            { return nil }
func (c *dgCallback) SpeechStarted(ssr *apiinterfaces.SpeechStartedResponse) error { return nil }
func (c *dgCallback) UtteranceEnd(ur *apiinterfaces.UtteranceEndResponse) error    { return nil }
func (c *dgCallback) Close(cr *apiinterfaces.CloseResponse) error                  { return nil }

func (c *dgCallback) Error(er *apiinterfaces.ErrorResponse) error {
	if er != nil {
		logger.Error("Deepgram error", zap.String("msg", er.ErrMsg))
	}
	return nil
}

func (c *dgCallback) UnhandledEvent(byData []byte) error { return nil }

// TranscribeStream consome um canal de áudio e retorna um canal de texto transcrito em tempo real.
func (c *STTClient) TranscribeStream(ctx context.Context, audioStream <-chan []byte) (<-chan string, error) {
	outStream := make(chan string, 100)

	opts := &interfaces.LiveTranscriptionOptions{
		Language:       "pt-BR",
		Model:          "nova-2",
		Punctuate:      true,
		Encoding:       "linear16",
		Channels:       1,
		SampleRate:     16000,
		InterimResults: true,
	}

	callback := &dgCallback{outStream: outStream}

	clientOptions := &interfaces.ClientOptions{}
	dgClient, err := listen.NewUsingCallback(ctx, c.apiKey, clientOptions, opts, callback)
	if err != nil {
		return nil, fmt.Errorf("failed to init deepgram client: %w", err)
	}

	bConnected := dgClient.Connect()
	if !bConnected {
		return nil, errors.New("failed to connect to Deepgram live API")
	}

	go func() {
		defer close(outStream)
		defer dgClient.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Info("TranscribeStream context canceled")
				return
			case audioData, ok := <-audioStream:
				if !ok {
					logger.Info("Audio stream closed")
					return
				}

				_, err := dgClient.Write(audioData)
				if err != nil {
					logger.Error("Error writing audio to deepgram", zap.Error(err))
					return
				}
			}
		}
	}()

	return outStream, nil
}
