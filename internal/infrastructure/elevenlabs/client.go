package elevenlabs

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/haguro/elevenlabs-go"
	"github.com/username/app-recrutamento-ia/internal/logger"
	"go.uber.org/zap"
)

// TTSClient implementa a interface domain.TTSClient usando a API da ElevenLabs.
type TTSClient struct {
	client  *elevenlabs.Client
	voiceID string
}

// NewTTSClient cria uma nova instância do TTSClient com a chave e ID de voz desejado.
func NewTTSClient(apiKey, voiceID string) (*TTSClient, error) {
	if apiKey == "" {
		return nil, errors.New("elevenlabs API key is required")
	}
	if voiceID == "" {
		return nil, errors.New("elevenlabs voice ID is required")
	}

	return &TTSClient{
		client:  elevenlabs.NewClient(context.Background(), apiKey, 30*time.Second),
		voiceID: voiceID,
	}, nil
}

// SynthesizeStream consome texto do LLM (streaming) e gera áudio via ElevenLabs API.
func (c *TTSClient) SynthesizeStream(ctx context.Context, textStream <-chan string) (<-chan []byte, error) {
	outStream := make(chan []byte, 100)

	go func() {
		defer close(outStream)

		var sentenceBuffer string

		for {
			select {
			case <-ctx.Done():
				logger.Info("TTS SynthesizeStream context canceled")
				return
			case textToken, ok := <-textStream:
				if !ok {
					// Stream de texto fechado. Sintetiza o que sobrou no buffer.
					if len(sentenceBuffer) > 0 {
						c.synthesizeAndSend(ctx, sentenceBuffer, outStream)
					}
					return
				}

				sentenceBuffer += textToken

				// Heurística simples para quebrar em sentenças (pontuação)
				if isPunctuation(textToken) && len(sentenceBuffer) > 10 {
					textToSynthesize := sentenceBuffer
					sentenceBuffer = "" // Limpa o buffer rapidamente para o próximo chunk

					c.synthesizeAndSend(ctx, textToSynthesize, outStream)
				}
			}
		}
	}()

	return outStream, nil
}

func (c *TTSClient) synthesizeAndSend(ctx context.Context, text string, outStream chan<- []byte) {
	req := elevenlabs.TextToSpeechRequest{
		Text:    text,
		ModelID: "eleven_multilingual_v2", // Suporta pt-BR com boa qualidade
		VoiceSettings: &elevenlabs.VoiceSettings{
			Stability:       0.5,
			SimilarityBoost: 0.7,
		},
	}

	logger.Debug("Synthesizing TTS chunk", zap.String("text", text))

	var audioBytes []byte

	// Adicionando Retry Logic robusta contra Rate Limiting e timeouts da ElevenLabs
	err := retry.Do(
		func() error {
			var err error
			audioBytes, err = c.client.TextToSpeech(c.voiceID, req)
			if err != nil {
				logger.Warn("ElevenLabs request failed, retrying...", zap.Error(err), zap.String("text", text))
				return err
			}
			return nil
		},
		retry.Attempts(3),
		retry.Delay(500*time.Millisecond),
		retry.MaxDelay(2*time.Second),
		retry.Context(ctx),
	)

	if err != nil {
		logger.Error("Failed to synthesize text after retries", zap.Error(err), zap.String("text", text))
		return
	}

	if len(audioBytes) > 0 {
		outStream <- audioBytes
	}
}

func isPunctuation(s string) bool {
	if len(s) == 0 {
		return false
	}
	lastChar := s[len(s)-1]
	return lastChar == '.' || lastChar == '?' || lastChar == '!' || lastChar == '\n'
}
