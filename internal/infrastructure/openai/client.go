package openai

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
	"github.com/username/app-recrutamento-ia/internal/domain"
	"github.com/username/app-recrutamento-ia/internal/logger"
	"go.uber.org/zap"
)

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

	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create completion stream: %w", err)
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
					logger.Error("OpenAI stream error", zap.Error(err))
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
