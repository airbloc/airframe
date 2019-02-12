package main

import (
	"github.com/airbloc/airframe/apiserver"
	"github.com/airbloc/airframe/database"
	"github.com/pkg/errors"
	"os"
	"runtime"

	"github.com/airbloc/logger"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	config := LoadConfig()

	log := logger.New("main")
	log.Info("Using %s configuration", config.Profile)

	db, err := initDatabase(config.Backend)
	if err != nil {
		log.Error("error: failed to initialize database", err)
		os.Exit(1)
	}

	s := apiserver.NewServer(db, config.Port)
	if err := s.Start(); err != nil {
		log.Error("failed to start server", err)
		os.Exit(1)
	}
}

func initDatabase(backendType string) (database.Database, error) {
	if backendType == "memory" {
		return database.NewInMemoryDatabase()
	}
	return nil, errors.Errorf("unknown backend: %s", backendType)
}
