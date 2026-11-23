package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"kpo_hw3_devenv/internal/storage/repository"
	"kpo_hw3_devenv/internal/storage/services"
	"kpo_hw3_devenv/internal/storage/transport"
)

func main() {
	if err := run(); err != nil {
		slog.Error("storage stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	addr := os.Getenv("STORAGE_ADDR")
	if addr == "" {
		addr = ":8081"
	}
	dataDir := os.Getenv("STORAGE_DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	db, err := repository.OpenSQLite(context.Background(), dataDir)
	if err != nil {
		return err
	}
	defer db.Close()
	archives, err := repository.NewArchiveStore(dataDir)
	if err != nil {
		return err
	}
	service := services.New(db, archives)
	mux := http.NewServeMux()
	transport.NewHandler(service).SetRoutes(mux)
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("storage starting", "addr", addr, "data_dir", dataDir)
	return server.ListenAndServe()
}
