package services

import (
	"os"

	"github.com/fandujar/baymax/pkg/providers"
	"github.com/rs/zerolog/log"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func NewSlackClient(appToken string, botToken string) *socketmode.Client {
	if appToken == "" {
		appToken = os.Getenv("SLACK_APP_TOKEN")
		if appToken == "" {
			return nil
		}
	}

	if botToken == "" {
		botToken = os.Getenv("SLACK_BOT_TOKEN")
		if botToken == "" {
			return nil
		}
	}

	api := slack.New(
		botToken,
		// slack.OptionDebug(true),
		slack.OptionAppLevelToken(appToken),
	)

	client := socketmode.New(
		api,
		// socketmode.OptionDebug(true),
	)

	return client
}

func RegisterHandlers(client *socketmode.Client) *socketmode.SocketmodeHandler {
	handler := socketmode.NewSocketmodeHandler(client)
	handler.Handle(socketmode.EventTypeConnecting, connectionHandler)
	handler.Handle(socketmode.EventTypeConnectionError, connectionHandler)
	handler.Handle(socketmode.EventTypeConnected, connectionHandler)
	handler.Handle(socketmode.EventType("hello"), connectionHandler)

	handler.HandleEvents(slackevents.AppMention, eventsAPIHandler)
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

func eventsAPIHandler(evt *socketmode.Event, client *socketmode.Client) {

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

	log.Info().Msgf("app_mention: %s\n", ev.Text)

	threadID := ev.ThreadTimeStamp + ev.Channel
	data := map[string]interface{}{
		"inputs": map[string]interface{}{
			"thread_id": threadID,
			"message":   ev.Text,
		},
		"response_mode": "streaming",
		"user":          ev.User,
	}

	provider := providers.NewDifyProvider(
		&providers.DifyProviderConfig{},
	)
	if err := provider.RunWorkflow(data); err != nil {
		log.Error().Err(err).Msg("failed to run workflow")
	}
}

func defaultHandler(e *socketmode.Event, client *socketmode.Client) {
	log.Printf("ignored %+v\n", e)
	client.Ack(*e.Request)
}
