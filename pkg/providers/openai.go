package providers

import (
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	*OpenAIProviderConfig
	Client *openai.Client
}

type OpenAIProviderConfig struct {
	BaseURL string
	Token   string
}

func NewOpenAIProvider(config *OpenAIProviderConfig) (*OpenAIProvider, error) {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}

	if config.Token == "" {
		config.Token = os.Getenv("OPENAI_API_KEY")
		if config.Token == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is required")
		}
	}

	client := openai.NewClient(config.Token)

	return &OpenAIProvider{
		OpenAIProviderConfig: config,
		Client:               client,
	}, nil
}
