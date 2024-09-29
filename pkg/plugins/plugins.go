package plugins

import (
	"fmt"
	"os"
	"plugin"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/sashabaranov/go-openai"
)

type Plugin interface {
	GetTools(log zerolog.Logger) []openai.Tool
	RunTool(log zerolog.Logger, toolName string, parameters string, messages []openai.ChatCompletionMessage, tools []openai.Tool) (string, error)
	RunEventLoop(log zerolog.Logger, natsClient *nats.Conn)
}

type PluginConfig struct {
	Name       string
	NatsClient *nats.Conn
}

func LoadPlugins() ([]Plugin, error) {
	plugins := []Plugin{}
	pluginsDir := os.Getenv("BAYMAX_PLUGINS_DIR")
	if pluginsDir == "" {
		pluginsDir = "."
	}

	pluginFiles := []string{}

	// for all .so files in the plugins directory
	// load the plugin and append it to the pluginFiles slice
	files, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".so") {
			pluginFiles = append(pluginFiles, fmt.Sprintf("%s/%s", pluginsDir, file.Name()))
		}
	}

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
