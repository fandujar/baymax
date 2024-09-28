package plugins

import (
	"github.com/nats-io/nats.go"
	"github.com/sashabaranov/go-openai"
)

type Plugin interface {
	New(config *PluginConfig) (Plugin, error)
	GetName() string
	GetTools() []openai.Tool
	RunEventLoop()
}

type PluginConfig struct {
	Name       string
	NatsClient *nats.Conn
}
