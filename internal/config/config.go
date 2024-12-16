package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		DBName   string
	}
	Redis struct {
		Host     string
		Port     int
		User     string
		Password string
	}
	RabbitMQ struct {
		Host string
		Port int
	}
	Server struct {
		Host string
		Port int
	}
	AWS struct {
		S3Bucket string
		Region   string
	}
}

func LoadConfig() *Config {
	viper.SetConfigName("config_dummy")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	var config Config

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}

	return &config
}
