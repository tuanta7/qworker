package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const envPrefix = "Q_WORKER"

type Config struct {
	ServerName string `envconfig:"SERVER_NAME" default:"worker"`
	ServerHost string `envconfig:"SERVER_HOST" default:"localhost"`
	ServerPort uint32 `envconfig:"SERVER_PORT" default:"8080"`
	AESSecret  string `envconfig:"AES_SECRET" default:"1234567887654321"`
	Logger     *LoggerConfig
	StartTLS   *StartTLSConfig
	Postgres   *PostgresConfig
	Redis      *RedisConfig
}

type LoggerConfig struct {
	Level      string `envconfig:"log_level" default:"info"`
	LogRequest bool   `envconfig:"log_request" default:"true"`
}

type RedisConfig struct {
	Sentinels  []string `envconfig:"REDIS_SENTINELS" default:"localhost:26379"`
	MasterName string   `envconfig:"REDIS_MASTER_NAME" default:"mymaster"`
	Password   string   `envconfig:"REDIS_PASSWORD" default:""`
	Database   int      `envconfig:"REDIS_DATABASE" default:"0"`
}

type StartTLSConfig struct {
	SkipVerify bool `envconfig:"SKIP_VERIFY" default:"false"`
}

type PostgresConfig struct {
	Host     string `envconfig:"POSTGRES_HOST" default:"localhost"`
	Port     uint32 `envconfig:"POSTGRES_PORT" default:"5432"`
	Username string `envconfig:"POSTGRES_USERNAME" default:"postgres"`
	Password string `envconfig:"POSTGRES_PASSWORD" default:"password"`
	Database string `envconfig:"POSTGRES_DATABASE" default:"qworker"`
}

func (p PostgresConfig) GetConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", p.Username, p.Password, p.Host, p.Port, p.Database)
}

func InitConfig() *Config {
	config := &Config{}

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
