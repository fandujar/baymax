package providers

import (
	"fmt"
	"os"
)

type OpenAIProvider struct {
	*OpenAIProviderConfig
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

	return &OpenAIProvider{config}, nil
}
