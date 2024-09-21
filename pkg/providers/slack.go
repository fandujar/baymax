package providers

import (
	"fmt"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type SlackProvider struct {
	*SlackProviderConfig
	Client *socketmode.Client
}

type SlackProviderConfig struct {
	AppToken string
	BotToken string
}

func NewSlackClient(appToken string, botToken string) (*socketmode.Client, error) {
	if appToken == "" {
		appToken = os.Getenv("SLACK_APP_TOKEN")
		if appToken == "" {
			return nil, fmt.Errorf("SLACK_APP_TOKEN is required")
		}
	}

	if botToken == "" {
		botToken = os.Getenv("SLACK_BOT_TOKEN")
		if botToken == "" {
			return nil, fmt.Errorf("SLACK_BOT_TOKEN is required")
		}
	}

	api := slack.New(
		botToken,
		slack.OptionAppLevelToken(appToken),
	)

	client := socketmode.New(
		api,
	)

	return client, nil
}

func NewSlackProvider(config *SlackProviderConfig) (*SlackProvider, error) {
	client, err := NewSlackClient(config.AppToken, config.BotToken)
	if err != nil {
		return nil, err
	}

	return &SlackProvider{config, client}, nil
}
