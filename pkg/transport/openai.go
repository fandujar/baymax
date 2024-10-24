package transport

import (
	"encoding/json"
	"os"

	"github.com/fandujar/baymax/pkg/plugins"
	"github.com/fandujar/baymax/pkg/services"
	"github.com/fandujar/baymax/pkg/subjects"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
)

type OpenAIHandler struct {
	Service *services.OpenAIService
	Tools   []openai.Tool
	Plugins []plugins.Plugin
}

func NewOpenAIHandler(service *services.OpenAIService, tools []openai.Tool, plugins []plugins.Plugin) *OpenAIHandler {
	return &OpenAIHandler{
		Service: service,
		Tools:   tools,
		Plugins: plugins,
	}
}

func (h *OpenAIHandler) RunEventLoop() {
	h.Service.NatsClient.Subscribe(subjects.SlackEvents, func(m *nats.Msg) {
		// Get the message and call the OpenAI API to get a response
		// Send the response to NATS using the subject SlackResponse

		ev := &ThreadMessage{}
		if err := json.Unmarshal(m.Data, ev); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal event")
			return
		}

		messages := []openai.ChatCompletionMessage{}

		systemMessage := os.Getenv("OPENAI_SYSTEM_MESSAGE")
		log.Debug().Str("system_message", systemMessage).Msg("system message")
		if systemMessage != "" {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    "system",
				Content: systemMessage,
			})
		}

		for _, message := range ev.Messages {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    "user",
				Content: message.Text,
			})
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    "user",
			Content: ev.Event.Text,
		})

		resp, err := h.Service.ChatCompletion(messages, h.Tools, h.Plugins)
		if err != nil {
			log.Error().Err(err).Msg("failed to get chat completion")
			return
		}

		ev.Event.Text = resp

		data, err := json.Marshal(ev)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal event")
			return
		}

		if err := h.Service.NatsClient.Publish(subjects.SlackResponse, data); err != nil {
			log.Error().Err(err).Msg("failed to publish message")
		}

	})
}
