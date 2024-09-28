package services

import (
	"context"
	"fmt"
	"os"

	"github.com/fandujar/baymax/pkg/plugins"
	"github.com/fandujar/baymax/pkg/providers"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
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

func (s *OpenAIService) ChatCompletion(messages []openai.ChatCompletionMessage, tools []openai.Tool, plugins []plugins.Plugin) (string, error) {
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

	var response string
	response = resp.Choices[0].Message.Content

	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		for _, toolCall := range resp.Choices[0].Message.ToolCalls {
			messages = append(messages, resp.Choices[0].Message)
			if toolCall.Function.Name == "MyNameIs" {
				messages = append(messages,
					openai.ChatCompletionMessage{
						Role:       openai.ChatMessageRoleTool,
						Content:    MyNameIs(),
						Name:       toolCall.Function.Name,
						ToolCallID: toolCall.ID,
					},
				)
			}

			for _, plugin := range plugins {
				for _, tool := range plugin.GetTools() {
					if toolCall.Function.Name == tool.Function.Name {
						log.Debug().Msgf("Running function %s", tool.Function.Name)

						pluginResponse, err := plugin.RunTool(toolCall.Function.Name, toolCall.Function.Arguments, messages, tools)
						if err != nil {
							pluginResponse = fmt.Sprintf("Error running function %s: %s", tool.Function.Name, err)
						}

						messages = append(messages,
							openai.ChatCompletionMessage{
								Role:       openai.ChatMessageRoleTool,
								Content:    pluginResponse,
								Name:       toolCall.Function.Name,
								ToolCallID: toolCall.ID,
							},
						)
					}
				}
			}
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

			response = resp.Choices[0].Message.Content
		}
	}

	return response, err

}

func MyNameIs() string {
	name := os.Getenv("BAYMAX_NAME")
	if name == "" {
		name = "Baymax"
	}

	return "My name is " + name
}
