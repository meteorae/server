package config

import (
	"errors"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func init() {
	configFileLocation, err := xdg.ConfigFile("meteorae/server.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get config file location")
	}

	viper.SetConfigName("server")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Dir(configFileLocation))
	viper.AutomaticEnv()

	// Disable debug messages by default
	viper.SetDefault("verbose", false)
	// Default server port
	viper.SetDefault("port", 42000) //nolint:gomnd
	// Enable Sentry reporting by default
	// The reason is that it's anonimized, and helps us a lot
	// to get feedback users might not submit or even know about.
	viper.SetDefault("crash_reporting", true)

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if ok := errors.As(err, &configFileNotFound); ok {
			err = viper.WriteConfigAs(configFileLocation)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to write config file")
			}
		} else {
			log.Fatal().Err(err).Msg("Failed to read config file")
		}
	}
}
