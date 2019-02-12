package main

import (
	"log"
	"os"
	"time"

	"github.com/airbloc/logger"

	"github.com/jinzhu/configor"
)

const (
	// EnvPrefix is used to override config attributes with environment variables.
	// e.g) ABL_PORT=8080 go run main.go
	EnvPrefix = "ABL"
)

// Config stores service configurations.
type Config struct {
	Profile string `default:"dev"`
	Port    int    `default:"8080"`

	// Kafka configurations
	Kafka struct {
		Topic      string
		BrokerList []string `yaml:"broker_list"`
	}

	// HTTP request configurations.
	RequestTimeout time.Duration `yaml:"request_timeout"`
	MaxConnections int           `yaml:"request_max_connections"`
}

// LoadConfigFrom loads default config.
func LoadConfigFrom(path string) *Config {
	config := Config{}
	cfgr := configor.New(&configor.Config{
		Environment: os.Getenv(EnvPrefix + "_PROFILE"),
		ENVPrefix:   EnvPrefix,
	})
	if err := cfgr.Load(&config, path); err != nil {
		log.Fatalf("failed to load config %s: %v\n", path, err)
	}

	// setup global logger accordingly.
	var writer logger.StandardWriter
	if config.Profile == "production" {
		writer = logger.NewStandardOutput(os.Stdout, "INFO", "*")
		writer.ColorsEnabled = false // Force JSON Output
	} else {
		writer = logger.NewStandardOutput(os.Stdout, "DEBUG", "*")
		writer.ColorsEnabled = true
	}
	logger.SetLogger(writer)
	return &config
}
