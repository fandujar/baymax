package services

import (
	"context"
	"os"

	"github.com/fandujar/baymax/pkg/providers"
	"github.com/nats-io/nats.go"
	"github.com/sashabaranov/go-openai"
)

type OpenAIService struct {
	NatsClient     *nats.Conn
	OpenAIProvider *providers.OpenAIProvider
}

func NewOpenAIService(p *providers.OpenAIProvider, nc *nats.Conn) *OpenAIService {
	return &OpenAIService{
		NatsClient:     nc,
		OpenAIProvider: p,
	}
}

func (s *OpenAIService) ChatCompletion(messages []openai.ChatCompletionMessage, tools []openai.Tool) (string, error) {
	tools = append(tools, openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "MyNameIs",
			Description: "Return the name of this assistant",
			Parameters:  nil,
		},
	})

	resp, err := s.OpenAIProvider.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT4oMini,
			Messages: messages,
			Tools:    tools,
		},
	)

	if err != nil || len(resp.Choices) == 0 {
		return "", err
	}

	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		for _, toolCall := range resp.Choices[0].Message.ToolCalls {
			if toolCall.Function.Name == "MyNameIs" {
				messages = append(messages, resp.Choices[0].Message)
				messages = append(messages,
					openai.ChatCompletionMessage{
						Role:       openai.ChatMessageRoleTool,
						Content:    MyNameIs(),
						Name:       toolCall.Function.Name,
						ToolCallID: toolCall.ID,
					},
				)

				resp, err := s.OpenAIProvider.Client.CreateChatCompletion(
					context.Background(),
					openai.ChatCompletionRequest{
						Model:    openai.GPT4oMini,
						Messages: messages,
						Tools:    tools,
					},
				)

				if err != nil || len(resp.Choices) == 0 {
					return "", err
				}

				return resp.Choices[0].Message.Content, nil
			}
		}
	}

	return resp.Choices[0].Message.Content, err

}

func MyNameIs() string {
	name := os.Getenv("BAYMAX_NAME")
	if name == "" {
		name = "Baymax"
	}

	return "My name is " + name
}
