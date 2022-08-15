package config

import (
	"errors"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const defaultPort = 42000

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
	viper.SetDefault("port", defaultPort)
	// Enable Sentry reporting by default
	// The reason is that it's anonymized, and helps us a lot
	// to get errors users might not submit or even know about.
	viper.SetDefault("crash_reporting", true)
	// Generate a random secret for JWT tokens using UUIDv4.
	// This should be secure enough, and random. Users can still
	// generate their own secret if they want.
	viper.SetDefault("jwt_secret", uuid.New().String())

	if configReadErr := viper.ReadInConfig(); configReadErr != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if ok := errors.As(configReadErr, &configFileNotFound); ok {
			configReadErr = viper.WriteConfigAs(configFileLocation)
			if configReadErr != nil {
				log.Fatal().Err(configReadErr).Msg("Failed to write config file")
			}
		} else {
			log.Fatal().Err(configReadErr).Msg("Failed to read config file")
		}
	}
}
