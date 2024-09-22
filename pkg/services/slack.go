package services

import (
	"github.com/fandujar/baymax/pkg/providers"
	"github.com/nats-io/nats.go"
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
