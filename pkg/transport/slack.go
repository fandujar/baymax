package transport

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func RegisterSlackHandlers(client *socketmode.Client, nc *nats.Conn) *socketmode.SocketmodeHandler {
	handler := socketmode.NewSocketmodeHandler(client)
	handler.Handle(socketmode.EventTypeConnecting, connectionHandler)
	handler.Handle(socketmode.EventTypeConnectionError, connectionHandler)
	handler.Handle(socketmode.EventTypeConnected, connectionHandler)
	handler.Handle(socketmode.EventType("hello"), connectionHandler)

	handler.HandleEvents(slackevents.AppMention, appMentionHandler(nc))
	handler.HandleDefault(defaultHandler)

	nc.Subscribe("slack.response", func(m *nats.Msg) {
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
			slack.MsgOptionText("Hello @"+ev.User+"!", false),
			slack.MsgOptionTS(ev.ThreadTimeStamp),
		)

		if _, _, _, err := client.SendMessage(
			ev.Channel,
			message,
		); err != nil {
			log.Error().Err(err).Msg("failed to send message")
		}
	})
	return handler
}

func connectionHandler(e *socketmode.Event, client *socketmode.Client) {
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

func appMentionHandler(nc *nats.Conn) func(*socketmode.Event, *socketmode.Client) {
	return func(evt *socketmode.Event, client *socketmode.Client) {
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
		if err := nc.Publish("slack.events", evJSON); err != nil {
			log.Error().Err(err).Msg("failed to publish event to NATS")
		}
	}
}

func defaultHandler(e *socketmode.Event, client *socketmode.Client) {
	log.Printf("ignored %+v\n", e)
	client.Ack(*e.Request)
}
