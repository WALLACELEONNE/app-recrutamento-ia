package openai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/sashabaranov/go-openai"
	"github.com/username/app-recrutamento-ia/internal/domain"
	"github.com/username/app-recrutamento-ia/internal/logger"
	"go.uber.org/zap"
)

// TTSClient implementa a interface domain.TTSClient usando a API da OpenAI.
type TTSClient struct {
	client *openai.Client
	voice  openai.SpeechVoice
}

// NewTTSClient cria uma nova instância do TTSClient com a chave da OpenAI.
func NewTTSClient(apiKey string, voice string) (*TTSClient, error) {
	if apiKey == "" {
		return nil, errors.New("openai API key is required")
	}

	v := openai.VoiceAlloy
	if voice != "" {
		v = openai.SpeechVoice(voice)
	}

	return &TTSClient{
		client: openai.NewClient(apiKey),
		voice:  v,
	}, nil
}

func isPunctuation(s string) bool {
	if len(s) == 0 {
		return false
	}
	lastChar := s[len(s)-1]
	return lastChar == '.' || lastChar == '?' || lastChar == '!' || lastChar == '\n'
}

// SynthesizeStream consome texto do LLM (streaming) e gera áudio via OpenAI API.
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
					if len(sentenceBuffer) > 0 {
						c.synthesizeAndSend(ctx, sentenceBuffer, outStream)
					}
					return
				}

				sentenceBuffer += textToken

				if isPunctuation(textToken) && len(sentenceBuffer) > 10 {
					textToSynthesize := sentenceBuffer
					sentenceBuffer = ""
					c.synthesizeAndSend(ctx, textToSynthesize, outStream)
				}
			}
		}
	}()

	return outStream, nil
}

func (c *TTSClient) synthesizeAndSend(ctx context.Context, text string, outStream chan<- []byte) {
	logger.Debug("Synthesizing TTS chunk", zap.String("text", text))

	var audioBytes []byte
	req := openai.CreateSpeechRequest{
		Model:          openai.TTSModel1,
		Voice:          c.voice,
		Input:          text,
		ResponseFormat: openai.SpeechResponseFormatOpus,
	}

	err := retry.Do(
		func() error {
			resp, err := c.client.CreateSpeech(ctx, req)
			if err != nil {
				logger.Warn("OpenAI TTS request failed, retrying...", zap.Error(err), zap.String("text", text))
				return err
			}
			defer resp.Close()

			audioBytes, err = io.ReadAll(resp)
			if err != nil {
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

// LLMClient implementa a interface domain.LLMClient usando OpenAI.
type LLMClient struct {
	client *openai.Client
}

// NewLLMClient cria uma nova instância do LLMClient com a chave da OpenAI.
func NewLLMClient(apiKey string) (*LLMClient, error) {
	if apiKey == "" {
		return nil, errors.New("openai API key is required")
	}

	return &LLMClient{
		client: openai.NewClient(apiKey),
	}, nil
}

// GenerateResponseStream envia o histórico e o prompt atual para a OpenAI e retorna um stream de tokens.
// Inclui lógica de retry robusta para tolerância a falhas na comunicação com a API externa.
func (c *LLMClient) GenerateResponseStream(ctx context.Context, systemPrompt string, history []domain.SessionTurn, currentTurn string) (<-chan string, error) {
	if systemPrompt == "" {
		return nil, errors.New("system prompt cannot be empty")
	}

	outStream := make(chan string, 100)

	// Constrói a lista de mensagens a partir do histórico
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
	}

	for _, turn := range history {
		role := openai.ChatMessageRoleUser
		if turn.Role == domain.RoleAI {
			role = openai.ChatMessageRoleAssistant
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: turn.Content,
		})
	}

	// Adiciona a fala atual do candidato
	if currentTurn != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: currentTurn,
		})
	}

	req := openai.ChatCompletionRequest{
		Model:     openai.GPT4o,
		Messages:  messages,
		Stream:    true,
		MaxTokens: 150, // Respostas curtas e conversacionais
	}

	var stream *openai.ChatCompletionStream

	// Lógica de retry para criação do stream (trata rate limits ou falhas de conexão iniciais)
	err := retry.Do(
		func() error {
			var err error
			stream, err = c.client.CreateChatCompletionStream(ctx, req)
			if err != nil {
				logger.Warn("OpenAI stream creation failed, retrying...", zap.Error(err))
				return err
			}
			return nil
		},
		retry.Attempts(3),
		retry.Delay(1*time.Second),
		retry.MaxDelay(5*time.Second),
		retry.Context(ctx),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create completion stream after retries: %w", err)
	}

	go func() {
		defer close(outStream)
		defer stream.Close()

		for {
			select {
			case <-ctx.Done():
				logger.Info("GenerateResponseStream context canceled")
				return
			default:
				response, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					return
				}
				if err != nil {
					logger.Error("OpenAI stream error during recv", zap.Error(err))
					return
				}

				if len(response.Choices) > 0 {
					content := response.Choices[0].Delta.Content
					if content != "" {
						outStream <- content
					}
				}
			}
		}
	}()

	return outStream, nil
}
