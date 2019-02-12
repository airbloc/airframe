package main

import (
	"fmt"
	"github.com/airbloc/airframe/database"
	"github.com/pkg/errors"
	"net/http"
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
	startServer(db, config.Port)
}

func initDatabase(backendType string) (database.Database, error) {
	if backendType == "memory" {
		return database.NewInMemoryDatabase()
	}
	return nil, errors.Errorf("unknown backend: %s", backendType)
}

func startServer(db database.Database, port int) {
	log := logger.New("webserver")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement me
	})

	// server listen
	log.Info("Server listening at http://localhost:%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != http.ErrServerClosed {
		log.Error("failed to start HTTP server", err)
	}
}
