package logging

import (
	"io"
	"io/ioutil"
	stdlog "log"
	"os"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}

	multi := zerolog.MultiLevelWriter(consoleWriter, newRollingFile())

	logger := zerolog.New(multi).With().Timestamp().Logger()

	log.Logger = logger

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if viper.GetBool("verbose") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Standard logging has no level and kind of sucks.
	stdlog.SetOutput(ioutil.Discard)
}

func newRollingFile() io.Writer {
	logPath, err := xdg.CacheFile("meteorae/logs/meteorae.log")
	if err != nil {
		log.Err(err).Msg("Failed to create log file")

		return nil
	}

	return &lumberjack.Logger{
		Filename: logPath,
	}
}
