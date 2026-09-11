package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func OpenSQLite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?_foreign_keys=on&_journal_mode=WAL&_locking_mode=EXCLUSIVE", path))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Starting application...")

	dbPath := os.Getenv("APP_DB")
	if dbPath == "" {
		dbPath = "statistics.db"
	}

	db, err := OpenSQLite(dbPath)

	if err != nil {
		logger.Error("Failed to open a database connection", "error", err)
		os.Exit(1)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping the database", "error", err)
		os.Exit(1)
	}

	logger.Info("Connected to the database!")

	if err := Migrate(db); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	logger.Info("Migrations applied")

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	store := NewSQLiteStore(db)
	mux := NewGameStatsServer(store)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		logger.Info("Starting server...")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
		}
	}()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	<-signalChannel

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Info("Initiating graceful shutdown...")

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Graceful shutdown failed", "error", err)
	}
}
