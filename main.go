package main

import "C"

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/getsentry/sentry-go"
	"github.com/meteorae/meteorae-server/agents"
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
	agents.InitAgentsManager()

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

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Err(err).Msg("The web server encountered an error")

			return
		}
	}()

	<-ctx.Done()

	stop()
	log.Info().Msg("Shutting down gracefully…")

	// Shutdown the web server and force-quit if it takes too long
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to gracefully shutdown web server")

		return
	}
}
