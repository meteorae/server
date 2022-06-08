package main

import "C"

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/99designs/gqlgen/cmd"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/getsentry/sentry-go"
	_ "github.com/meteorae/meteorae-server/config"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	_ "github.com/meteorae/meteorae-server/logging"
	_ "github.com/meteorae/meteorae-server/resolvers/all"
	"github.com/meteorae/meteorae-server/server"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var serverShutdownTimeout = 10 * time.Second

func main() {
	defer ants.Release()

	enable_reporting := viper.GetBool("crash_reporting")

	if enable_reporting {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:     "https://9ad21ea087cb4de1a5d2cfb6f36d354b@o725130.ingest.sentry.io/61632320",
			Debug:   viper.GetBool("verbose"),
			Release: fmt.Sprint("meteorae-server@%v", helpers.Version),
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize Sentry")

			return
		}

		defer sentry.Flush(serverShutdownTimeout)
	}

	log.Info().Msgf("Starting Meteorae %s", helpers.Version)
	log.Info().Msgf("Build Date: %s", helpers.BuildDate)
	log.Info().Msgf("Git Commit: %s", helpers.GitCommit)
	log.Info().Msgf("Go Version: %s", helpers.GoVersion)
	log.Info().Msgf("OS / Arch: %s", helpers.OsArch)

	// Initialize the database
	err := database.NewDatabase()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize database")

		return
	}

	vips.Startup(nil)
	defer vips.Shutdown()

	srv, err := server.GetWebServer()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize web server")

		return
	}

	go func() {
		log.Info().Msg("Starting the web server…")

		if err := srv.ListenAndServe(); !errors.Is(err, nil) {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Err(err).Msg("The web server encountered an error")
			} else {
				log.Info().Msg("The web server stopped")
			}
		}
	}()

	c := make(chan os.Signal, 1)
	// TODO: Handle SIGKILL, SIGQUIT and SIGTERM
	signal.Notify(c, os.Interrupt)

	// Block until we get our signal
	<-c

	log.Info().Msg("Received a SIGINT signal, shutting down gracefully…")

	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		log.Err(err).Msg("The web server encountered an error while shutting down")
	}

	log.Info().Msg("Shutting down…")
}
