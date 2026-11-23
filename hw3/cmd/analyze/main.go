package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"kpo_hw3_devenv/internal/analyze/repository"
	"kpo_hw3_devenv/internal/analyze/services"
	"kpo_hw3_devenv/internal/analyze/transport"
)

func main() {
	if err := run(); err != nil {
		slog.Error("analyze stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	addr := os.Getenv("ANALYZE_ADDR")
	if addr == "" {
		addr = ":8082"
	}
	storageURL := os.Getenv("STORAGE_URL")
	if storageURL == "" {
		storageURL = "http://localhost:8081"
	}

	storage := repository.NewStorage(storageURL, &http.Client{Timeout: 5 * time.Second})
	service := services.New(storage)
	mux := http.NewServeMux()
	transport.NewHandler(service).SetRoutes(mux)
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("analyze starting", "addr", addr, "storage_url", storageURL)
	return server.ListenAndServe()
}
