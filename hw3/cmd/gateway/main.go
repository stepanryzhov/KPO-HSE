package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"kpo_hw3_devenv/internal/gateway/repository"
	"kpo_hw3_devenv/internal/gateway/services"
	"kpo_hw3_devenv/internal/gateway/transport"
)

const defaultMaxZIPBytes int64 = 10 << 20

func main() {
	if err := run(); err != nil {
		slog.Error("gateway stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	addr := os.Getenv("GATEWAY_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	storageURL := os.Getenv("STORAGE_URL")
	if storageURL == "" {
		storageURL = "http://localhost:8081"
	}
	analyzeURL := os.Getenv("ANALYZE_URL")
	if analyzeURL == "" {
		analyzeURL = "http://localhost:8082"
	}
	maxZIPBytes := defaultMaxZIPBytes
	if raw := os.Getenv("GATEWAY_MAX_ZIP_BYTES"); raw != "" {
		limit, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || limit <= 0 {
			return fmt.Errorf("GATEWAY_MAX_ZIP_BYTES must be a positive integer: %q", raw)
		}
		maxZIPBytes = limit
	}

	client := &http.Client{Timeout: 5 * time.Second}
	storage := repository.NewStorage(storageURL, client)
	analyze := repository.NewAnalyze(analyzeURL, client)
	service := services.New(storage, analyze)
	mux := http.NewServeMux()
	transport.NewHandler(service, maxZIPBytes).SetRoutes(mux)
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("gateway starting", "addr", addr, "storage_url", storageURL, "analyze_url", analyzeURL, "max_zip_bytes", maxZIPBytes)
	return server.ListenAndServe()
}
