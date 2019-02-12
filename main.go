package main

import (
	"github.com/airbloc/airframe/apiserver"
	"github.com/airbloc/airframe/database"
	"github.com/airbloc/airframe/rpcserver"
	"github.com/pkg/errors"
	"os"
	"runtime"
	"sync"

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

	api := apiserver.New(db, config.Port, config.Profile == "dev")
	rpc := rpcserver.New(db, config.RpcPort, config.Profile == "dev")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		if err := api.Start(); err != nil {
			log.Error("failed to start api server", err)
			os.Exit(1)
		}
		wg.Done()
	}()

	go func() {
		if err := rpc.Start(); err != nil {
			log.Error("failed to start rpc server", err)
			os.Exit(1)
		}
		wg.Done()
	}()
	wg.Wait()
}

func initDatabase(backendType string) (database.Database, error) {
	if backendType == "memory" {
		return database.NewInMemoryDatabase()
	}
	return nil, errors.Errorf("unknown backend: %s", backendType)
}
