package main

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/airbloc/logger"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	config := LoadConfigFrom("resources/config.yml")

	log := logger.New("main")
	log.Info("Using %s configuration", config.Profile)

	testWebServer(config)
}

func testWebServer(config *Config) {
	log := logger.New("webserver")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!")
	})

	// server listen
	log.Info("Server listening at http://localhost:%d", config.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), mux); err != http.ErrServerClosed {
		log.Error("failed to start HTTP server", err)
	}
}
