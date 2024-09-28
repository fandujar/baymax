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

type ThreadMessage struct {
	Event    *slackevents.AppMentionEvent `json:"event"`
	Messages []slack.Message              `json:"messages"`
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
		var ev *ThreadMessage
		if err := json.Unmarshal(m.Data, &ev); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal event")
			return
		}

		if ev.Event.ThreadTimeStamp == "" {
			log.Debug().Msg("thread timestamp is empty")
			ev.Event.ThreadTimeStamp = ev.Event.TimeStamp
		}

		message := slack.MsgOptionCompose(
			slack.MsgOptionText(ev.Event.Text, false),
			slack.MsgOptionTS(ev.Event.ThreadTimeStamp),
		)

		if _, _, _, err := h.Service.SlackProvider.Client.SendMessage(
			ev.Event.Channel,
			message,
		); err != nil {
			log.Error().Err(err).Msgf("failed to send message: %s", err)
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

	var messages []slack.Message
	var err error
	if ev.ThreadTimeStamp != "" {
		// If inside a thread, get all messages in the thread to pass as context
		messages, err = h.Service.GetAllMessagesFromThread(ev.Channel, ev.ThreadTimeStamp)
		if err != nil {
			log.Error().Err(err).Msg("failed to get messages from thread")
			return
		}
	}

	threadMessage := &ThreadMessage{
		Event:    ev,
		Messages: messages,
	}

	threadMessageJSON, err := json.Marshal(threadMessage)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal event")
		return
	}

	if err := h.Service.NatsClient.Publish(subjects.SlackEvents, threadMessageJSON); err != nil {
		log.Error().Err(err).Msg("failed to publish event to NATS")
	}
}

func defaultHandler(e *socketmode.Event, client *socketmode.Client) {
	log.Printf("ignored %+v\n", e)
	client.Ack(*e.Request)
}
