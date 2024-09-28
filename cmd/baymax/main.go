package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/fandujar/baymax/pkg/plugins"
	"github.com/fandujar/baymax/pkg/providers"
	"github.com/fandujar/baymax/pkg/services"
	"github.com/fandujar/baymax/pkg/transport"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
)

func main() {
	// Configure the logger level and format
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = time.RFC3339Nano

	log.Info().Msg("starting baymax")

	var healthCheckServer *http.Server
	s := make(chan os.Signal, 1)
	shutdown := make(chan bool, 1)
	signal.Notify(s, os.Interrupt)

	go func() {
		signal := <-s
		log.Info().Msgf("Received signal: %v", signal)
		shutdown <- true
	}()

	healthCheck := chi.NewRouter()
	healthCheck.Get("/liveness", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	healthCheck.Get("/readiness", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	healthCheckServer = &http.Server{
		Addr:    ":8081",
		Handler: healthCheck,
	}

	go func() {
		log.Info().Msg("Starting health check server")
		if err := healthCheckServer.ListenAndServe(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start health check server")
		}
	}()

	// Start NATS server
	natsProvider, err := providers.NewNatsProvider(
		&providers.NatsProviderConfig{},
	)

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create nats provider")
	}

	go natsProvider.RunServer()

	nc, err := natsProvider.NewClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create nats client")
	}

	slackProvider, err := providers.NewSlackProvider(
		&providers.SlackProviderConfig{
			AppToken: os.Getenv("SLACK_APP_TOKEN"),
			BotToken: os.Getenv("SLACK_BOT_TOKEN"),
		},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create slack provider")
	}

	openAIProvider, err := providers.NewOpenAIProvider(
		&providers.OpenAIProviderConfig{
			Token: os.Getenv("OPENAI_API_KEY"),
		},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create openai provider")
	}

	// Load Plugins from plugins directory
	plugins, err := plugins.LoadPlugins()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load plugins")
	}

	// Load Tools from plugins
	tools := []openai.Tool{}
	for _, plugin := range plugins {
		tools = append(tools, plugin.GetTools()...)
	}

	// Start Services
	slackService := services.NewSlackService(slackProvider, nc)
	openAIService := services.NewOpenAIService(openAIProvider, nc)

	// Start Transport
	slackHandler := transport.NewSlackHandler(slackService)
	openAIHandler := transport.NewOpenAIHandler(openAIService, tools, plugins)
	slackHandler.RunEventLoop()
	openAIHandler.RunEventLoop()
	for _, plugin := range plugins {
		plugin.RunEventLoop(nc)
	}

	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := healthCheckServer.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to shutdown health check server")
	}

	natsProvider.StopServer()
	nc.Close()

}
