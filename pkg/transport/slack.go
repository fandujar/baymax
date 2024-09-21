package transport

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
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
	// handler.HandleEvents(, eventsHandler)

	handler.HandleDefault(defaultHandler)

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
			log.Info().Msgf("ignored %+v\n", evt)
			return
		}

		client.Ack(*evt.Request)

		ev, ok := eventsAPIEvent.InnerEvent.Data.(*slackevents.AppMentionEvent)
		if !ok {
			log.Info().Msgf("ignored %+v\n", evt)
			return
		}

		log.Debug().Msgf("app_mention: %s\n", ev.Text)

		threadID := ev.ThreadTimeStamp + ev.Channel
		data := map[string]interface{}{
			"inputs": map[string]interface{}{
				"thread_id": threadID,
				"message":   ev.Text,
			},
			"response_mode": "streaming",
			"user":          ev.User,
		}

		log.Debug().Msgf("publishing to nats: %v\n", data)
		nc.Publish("not.classified", []byte(fmt.Sprintf("%v", data)))
	}
}

func defaultHandler(e *socketmode.Event, client *socketmode.Client) {
	log.Printf("ignored %+v\n", e)
	client.Ack(*e.Request)
}
