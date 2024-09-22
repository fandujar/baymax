package services

import (
	"github.com/fandujar/baymax/pkg/providers"
	"github.com/nats-io/nats.go"
	"github.com/slack-go/slack"
)

type SlackService struct {
	SlackProvider *providers.SlackProvider
	NatsClient    *nats.Conn
}

func NewSlackService(slackProvider *providers.SlackProvider, natsClient *nats.Conn) *SlackService {
	return &SlackService{
		SlackProvider: slackProvider,
		NatsClient:    natsClient,
	}
}

func (s *SlackService) GetAllMessagesFromThread(channel, threadTimestamp string) ([]slack.Message, error) {
	messages, _, _, err := s.SlackProvider.Client.GetConversationReplies(
		&slack.GetConversationRepliesParameters{
			ChannelID: channel,
			Timestamp: threadTimestamp,
			Limit:     100,
		},
	)
	if err != nil {
		return nil, err
	}

	return messages, nil
}
