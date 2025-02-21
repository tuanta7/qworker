package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const envPrefix = "Q_WORKER"

type Config struct {
	ServerName string `envconfig:"SERVER_NAME" default:"worker"`
	ServerHost string `envconfig:"SERVER_HOST" default:"localhost"`
	ServerPort uint32 `envconfig:"SERVER_PORT" default:"8080"`
	Postgres   *PostgresConfig
	Redis      *RedisConfig
}

type PostgresConfig struct {
	Host     string `envconfig:"POSTGRES_HOST" default:"localhost"`
	Port     uint32 `envconfig:"POSTGRES_PORT" default:"5432"`
	Username string `envconfig:"POSTGRES_USERNAME" default:"postgres"`
	Password string `envconfig:"POSTGRES_PASSWORD" default:"password"`
	Database string `envconfig:"POSTGRES_DATABASE" default:"worker"`
}

type RedisConfig struct {
	Sentinels []string `envconfig:"REDIS_SENTINELS" default:"localhost:26379"`
	Master    string   `envconfig:"REDIS_MASTER" default:"mymaster"`
	Password  string   `envconfig:"REDIS_PASSWORD" default:"password"`
}

func NewConfig() *Config {
	var config = &Config{}

	err := godotenv.Load()
	if err != nil {
		log.Printf("[Warning] config - init - godotenv.Load: %v", err)
	}

	err = envconfig.Process(envPrefix, config)
	if err != nil {
		log.Fatalf("config - init - envconfig.Process: %v", err)
	}

	return config
}
