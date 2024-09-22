package transport

import (
	"encoding/json"

	"github.com/fandujar/baymax/pkg/services"
	"github.com/fandujar/baymax/pkg/subjects"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack/slackevents"
)

type OpenAIHandler struct {
	Service *services.OpenAIService
}

func NewOpenAIHandler(service *services.OpenAIService) *OpenAIHandler {
	return &OpenAIHandler{
		Service: service,
	}
}

func (h *OpenAIHandler) RunEventLoop() {
	h.Service.NatsClient.Subscribe(subjects.SlackEvents, func(m *nats.Msg) {
		// Get the message and call the OpenAI API to get a response
		// Send the response to NATS using the subject SlackResponse

		ev := &slackevents.AppMentionEvent{}
		if err := json.Unmarshal(m.Data, ev); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal event")
			return
		}

		resp, err := h.Service.ChatCompletion(ev.Text)
		if err != nil {
			log.Error().Err(err).Msg("failed to get chat completion")
			return
		}

		ev.Text = resp

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
