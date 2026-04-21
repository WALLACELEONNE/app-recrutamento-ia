package elevenlabs

import (
	"context"
	"errors"
	"time"

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
// NOTA: Para um streaming real e contínuo (texto parcial -> áudio parcial), a ElevenLabs recomenda
// usar a API WebSocket. O SDK atual `haguro/elevenlabs-go` suporta chamadas HTTP (TextToSpeech).
// Para manter a simplicidade inicial, vamos agrupar pequenas sentenças e chamar a API de Streaming HTTP.
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
				// Em produção, isso deve ser mais robusto para não cortar no meio de siglas, etc.
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

	// TextToSpeech retorna o array de bytes do áudio (MP3 ou PCM dependendo da config)
	// Para LiveKit, o ideal seria PCM, mas podemos configurar o header de accept no client interno.
	// Por padrão retorna mp3.
	audioBytes, err := c.client.TextToSpeech(c.voiceID, req)
	if err != nil {
		logger.Error("Failed to synthesize text", zap.Error(err), zap.String("text", text))
		return
	}

	if len(audioBytes) > 0 {
		// Envia os bytes gerados para o canal de saída (que será lido pelo LiveKit)
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
