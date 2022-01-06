package main

import "C"

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/server"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/gographics/imagick.v3/imagick"
)

var serverShutdownTimeout = 10 * time.Second

func main() {
	defer ants.Release()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msgf("Starting Meteorae %s", helpers.Version)
	log.Info().Msgf("Build Date: %s", helpers.BuildDate)
	log.Info().Msgf("Git Commit: %s", helpers.GitCommit)
	log.Info().Msgf("Go Version: %s", helpers.GoVersion)
	log.Info().Msgf("OS / Arch: %s", helpers.OsArch)

	imagick.Initialize()
	defer imagick.Terminate()
	imageMagickVersion, _ := imagick.GetVersion()
	log.Info().Msgf("ImageMagick Version: %s", imageMagickVersion)

	// Initialize the database
	err := database.GetDatabase(log.Logger)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize database")

		return
	}

	var sqliteVersion interface{}

	rows, rowsErr := database.DB.Raw("SELECT sqlite_version()").Rows()
	if rowsErr != nil {
		log.Error().Err(rowsErr)
	}

	if rows.Err() != nil {
		log.Error().Err(rows.Err())
	}
	defer rows.Close()
	rows.Next()

	err = rows.Scan(&sqliteVersion)
	if err != nil {
		log.Error().Err(err)
	}

	log.Info().Msgf("SQLite Version: %s", sqliteVersion)

	var loadedSqliteExtensions []string

	rows, rowsErr = database.DB.Raw("PRAGMA compile_options").Rows()
	if rowsErr != nil {
		log.Error().Err(rowsErr)
	}

	if rows.Err() != nil {
		log.Error().Err(rows.Err())
	}
	defer rows.Close()

	for rows.Next() {
		var extensionRow interface{}

		err = rows.Scan(&extensionRow)
		if err != nil {
			log.Error().Err(err)
		}

		loadedSqliteExtensions = append(loadedSqliteExtensions, extensionRow.(string))
	}

	log.Info().Msgf("SQLite build information: %s", strings.Join(loadedSqliteExtensions, " "))

	log.Info().Msg("Checking for database migrations…")

	err = database.Migrate()
	if errors.Is(err, nil) {
		log.Error().Msgf("Could not migrate: %v", err)

		return
	}

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
