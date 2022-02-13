package config

import (
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

	viper.SetDefault("verbose", false)
	viper.SetDefault("port", 42000) //nolint:gomnd

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err = viper.WriteConfigAs(configFileLocation)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to write config file")
			}
		} else {
			log.Fatal().Err(err).Msg("Failed to read config file")
		}
	}
}
