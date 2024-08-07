package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/fandujar/baymax/pkg/services"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	slackClient := services.NewSlackClient("", "")
	handler := services.RegisterHandlers(slackClient)
	go func() {
		if err := handler.RunEventLoop(); err != nil {
			log.Fatal().Err(err).Msg("Failed to run event loop")
		}
	}()

	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := healthCheckServer.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to shutdown health check server")
	}

}
