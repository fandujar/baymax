package plugins

import (
	"github.com/nats-io/nats.go"
	"github.com/sashabaranov/go-openai"
)

type Plugin interface {
	New(config *PluginConfig) (Plugin, error)
	GetName() string
	GetTools() []openai.Tool
	RunTool(toolName string, messages []openai.ChatCompletionMessage, tools []openai.Tool) (string, error)
	RunEventLoop()
}

type PluginConfig struct {
	Name       string
	NatsClient *nats.Conn
}
