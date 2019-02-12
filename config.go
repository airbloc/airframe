package main

import (
	"github.com/spf13/pflag"
	"os"

	"github.com/airbloc/logger"
)

// Config stores service configurations.
type Config struct {
	Profile string `default:"dev"`
	Port    int    `default:"8080"`
	Backend string `default:"memory"`
}

// LoadConfigFrom loads default config.
func LoadConfig() *Config {
	isDev := pflag.BoolP("dev", "d", false, "Enable development mode.")
	port := pflag.IntP("port", "p", 8080, "Port of API server.")
	backend := pflag.StringP("backend", "b", "memory", "Backend type. [memory|leveldb|dynamodb]")
	pflag.Parse()

	// setup config from flag
	config := &Config{
		Port:    *port,
		Backend: *backend,
	}
	if *isDev {
		config.Profile = "dev"
	} else {
		config.Profile = "production"
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
	return config
}
