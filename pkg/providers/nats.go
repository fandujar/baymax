package providers

import (
	"time"

	natsServer "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type NatsProvider struct {
	*NatsProviderConfig
	Server *natsServer.Server
}

type NatsProviderConfig struct {
	natsServer.Options
}

func NewNatsProvider(config *NatsProviderConfig) (*NatsProvider, error) {
	if config.Options.Port == 0 {
		config.Options.Port = 4222
	}

	if config.Options.Host == "" {
		config.Options.Host = "localhost"
	}

	server := natsServer.New(&config.Options)

	return &NatsProvider{
		config,
		server,
	}, nil
}

func (n *NatsProvider) RunServer() {
	log.Info().Msg("starting nats server")
	n.Server.Start()
}

func (n *NatsProvider) StopServer() {
	log.Info().Msg("stopping nats server")
	n.Server.Shutdown()
}

func (n *NatsProvider) NewClient() (*nats.Conn, error) {
	log.Info().Msg("creating nats client")
	c, err := nats.Connect(
		n.Options.Host,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(1*time.Second),
	)

	if err != nil {
		return nil, err
	}

	return c, nil
}
