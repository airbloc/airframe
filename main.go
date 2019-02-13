package main

import (
	"github.com/airbloc/airframe/apiserver"
	"github.com/airbloc/airframe/database"
	"github.com/airbloc/airframe/rpcserver"
	"github.com/airbloc/logger"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"runtime"
	"time"
)

type Server interface {
	Start() error
	Stop()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	config := LoadConfig()

	log := logger.New("main")
	log.Info("Using {} configuration", config.Profile)

	db, err := initDatabase(config.Backend)
	if err != nil {
		log.Error("error: failed to initialize database", err)
		os.Exit(1)
	}

	// start API and RPC server
	servers := map[string]Server{
		"API": apiserver.New(db, config.Port, config.Profile == "dev"),
		"RPC": rpcserver.New(db, config.RpcPort, config.Profile == "dev"),
	}
	for name, server := range servers {
		go func() {
			if err := server.Start(); err != nil {
				log.Error("failed to start {} server", err, name)
				os.Exit(1)
			}
		}()
		time.Sleep(1 * time.Second)
		log.Info("{} server started", name)
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	for _, server := range servers {
		server.Stop()
	}
	log.Info("bye")
}

func initDatabase(backendType string) (database.Database, error) {
	if backendType == "memory" {
		return database.NewInMemoryDatabase()
	}
	return nil, errors.Errorf("unknown backend: %s", backendType)
}
