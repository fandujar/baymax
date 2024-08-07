package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type DifyProvider struct {
	*DifyProviderConfig
}

type DifyProviderConfig struct {
	BaseURL string
	Token   string
}

func NewDifyProvider(config *DifyProviderConfig) *DifyProvider {
	if config.Token == "" {
		config.Token = os.Getenv("DIFY_TOKEN")
		if config.Token == "" {
			return nil
		}
	}

	if config.BaseURL == "" {
		config.BaseURL = os.Getenv("DIFY_BASE_URL")
		if config.BaseURL == "" {
			config.BaseURL = "https://api.dify.ai"
		}
	}

	return &DifyProvider{
		DifyProviderConfig: config,
	}
}

func (p *DifyProvider) GetToken() string {
	return p.Token
}

func (p *DifyProvider) RunWorkflow(data interface{}) error {
	url := p.BaseURL + "/v1/workflows/run"

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+p.Token)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to run workflow: status code %d", response.StatusCode)
	}

	return nil
}
