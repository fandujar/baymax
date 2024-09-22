package transport

import (
	"encoding/json"

	"github.com/fandujar/baymax/pkg/services"
	"github.com/fandujar/baymax/pkg/subjects"
	"github.com/nats-io/nats.go"

	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type SlackHandler struct {
	Service *services.SlackService
}

func NewSlackHandler(service *services.SlackService) *SlackHandler {
	return &SlackHandler{
		Service: service,
	}
}

func (h *SlackHandler) RunEventLoop() {
	handler := h.RegisterSlackHandlers()
	go func() {
		if err := handler.RunEventLoop(); err != nil {
			log.Error().Err(err).Msg("failed to run event loop")
		}
	}()

	h.Service.NatsClient.Subscribe(subjects.SlackResponse, func(m *nats.Msg) {
		log.Debug().Msgf("Received a message: %s", string(m.Data))
		var ev *slackevents.AppMentionEvent
		if err := json.Unmarshal(m.Data, &ev); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal event")
			return
		}

		if ev.ThreadTimeStamp == "" {
			log.Debug().Msg("thread timestamp is empty")
			ev.ThreadTimeStamp = ev.TimeStamp
		}

		message := slack.MsgOptionCompose(
			slack.MsgOptionText(ev.Text, false),
			slack.MsgOptionTS(ev.ThreadTimeStamp),
		)

		if _, _, _, err := h.Service.SlackProvider.Client.SendMessage(
			ev.Channel,
			message,
		); err != nil {
			log.Error().Err(err).Msg("failed to send message")
		}
	})
}

func (h *SlackHandler) RegisterSlackHandlers() *socketmode.SocketmodeHandler {
	handler := socketmode.NewSocketmodeHandler(h.Service.SlackProvider.Client)
	handler.Handle(socketmode.EventTypeConnecting, h.connectionHandler)
	handler.Handle(socketmode.EventTypeConnectionError, h.connectionHandler)
	handler.Handle(socketmode.EventTypeConnected, h.connectionHandler)
	handler.Handle(socketmode.EventType("hello"), h.connectionHandler)

	handler.HandleEvents(slackevents.AppMention, h.appMentionHandler)
	handler.HandleDefault(defaultHandler)

	return handler
}

func (h *SlackHandler) connectionHandler(e *socketmode.Event, client *socketmode.Client) {
	switch e.Type {
	case socketmode.EventTypeConnecting:
		log.Info().Msg("connecting to slack with socket mode...")
	case socketmode.EventTypeConnectionError:
		log.Error().Msg("connection error.")
	case socketmode.EventTypeConnected:
		log.Info().Msg("connected to slack with socket mode.")
	case socketmode.EventType("hello"):
		log.Info().Msg("hello received.")
	default:
		log.Info().Msgf("ignored %+v\n", e)
	}
}

func (h *SlackHandler) appMentionHandler(evt *socketmode.Event, client *socketmode.Client) {
	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		log.Debug().Msgf("ignored %+v\n", evt)
		return
	}

	client.Ack(*evt.Request)

	ev, ok := eventsAPIEvent.InnerEvent.Data.(*slackevents.AppMentionEvent)
	if !ok {
		log.Debug().Msgf("ignored %+v\n", evt)
		return
	}

	evJSON, err := json.Marshal(ev)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal event")
		return
	}

	log.Debug().Msgf("Received a message: %s", string(evJSON))
	if err := h.Service.NatsClient.Publish(subjects.SlackEvents, evJSON); err != nil {
		log.Error().Err(err).Msg("failed to publish event to NATS")
	}
}

func defaultHandler(e *socketmode.Event, client *socketmode.Client) {
	log.Printf("ignored %+v\n", e)
	client.Ack(*e.Request)
}
