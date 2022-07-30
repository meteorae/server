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

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/getsentry/sentry-go"
	_ "github.com/meteorae/meteorae-server/config"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	_ "github.com/meteorae/meteorae-server/logging"
	"github.com/meteorae/meteorae-server/scanners"
	"github.com/meteorae/meteorae-server/server"
	"github.com/meteorae/meteorae-server/tasks"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var serverShutdownTimeout = 10 * time.Second

func main() {
	// If crash reporting is enabled, initialize Sentry.
	// We do this first so that we can capture any panics that occur during startup.
	enableReporting := viper.GetBool("crash_reporting")
	if enableReporting {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:     "https://9ad21ea087cb4de1a5d2cfb6f36d354b@o725130.ingest.sentry.io/61632320",
			Debug:   viper.GetBool("verbose"),
			Release: fmt.Sprintf("meteorae-server@%s", helpers.Version),
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize Sentry")

			return
		}
	}

	// Release all the workers when the server shuts down
	defer ants.Release()

	// Give the user some basic information about which version of Meteorae is running and on what
	log.Info().Msgf("Starting Meteorae %s", helpers.Version)
	log.Info().Msgf("Build Date: %s", helpers.BuildDate)
	log.Info().Msgf("Git Commit: %s", helpers.GitCommit)
	log.Info().Msgf("Go Version: %s", helpers.GoVersion)
	log.Info().Msgf("OS / Arch: %s", helpers.OsArch)

	// Initialize the database
	err := database.NewDatabase(log.Logger)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize database")

		return
	}

	// Initialize VIPS for image processing
	vips.Startup(nil)
	defer vips.Shutdown()

	scanners.InitScannersManager()

	// Initialize the task queue
	err = tasks.StartTaskQueues()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize task queues")

		return
	}
	defer tasks.StopTaskQueues()

	// Initialize the web server and all its components
	srv, err := server.GetWebServer()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize web server")

		return
	}

	log.Info().Msg("Starting the web server…")

	if err := srv.ListenAndServe(); !errors.Is(err, nil) {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Err(err).Msg("The web server encountered an error")
		} else {
			log.Info().Msg("The web server stopped")
		}
	}

	c := make(chan os.Signal, 1)
	// TODO: Handle SIGKILL, SIGQUIT and SIGTERM
	signal.Notify(c, os.Interrupt)

	// Block until we get our signal
	<-c

	log.Info().Msg("Received a SIGINT signal, shutting down gracefully…")

	// Shutdown the web server and force-quit if it takes too long
	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		log.Err(err).Msg("The web server encountered an error while shutting down")
	}
}
