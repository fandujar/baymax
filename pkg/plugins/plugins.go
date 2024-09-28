package plugins

import (
	"fmt"
	"plugin"

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

func LoadPlugins() ([]Plugin, error) {
	plugins := []Plugin{}
	pluginFiles := []string{"plugin.so"}

	// Load plugins from plugins directory
	for _, pluginFile := range pluginFiles {
		raw, err := plugin.Open(pluginFile)
		if err != nil {
			return nil, err
		}

		p, err := raw.Lookup("Plugin")
		if err != nil {
			return nil, err
		}

		plugin, ok := p.(Plugin)
		if !ok {
			return nil, fmt.Errorf("Plugin does not implement the Plugin interface")
		}

		plugins = append(plugins, plugin)
	}

	return plugins, nil
}
