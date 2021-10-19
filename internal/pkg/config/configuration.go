package config

import (
	"log"

	"github.com/spf13/viper"
)

var Config *Configuration

type Configuration struct {
	Server   ServerConfiguration
	Algod	AlgodConfiguration
}

type ServerConfiguration struct {
	Port   string
	Secret string
	Mode   string
}

type AlgodConfiguration struct {
	Address string
	Token string
}

func Setup(configPath string) {
	var configuration *Configuration

	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	Config = configuration
}

func GetConfig() *Configuration {
	return Config
}