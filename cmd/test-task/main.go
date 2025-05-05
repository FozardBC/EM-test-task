package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"test-task/internal/api"
	"test-task/internal/config"
	"test-task/internal/logger"
	"time"

	"test-task/internal/storage/postgres"

	"github.com/golang-migrate/migrate/v3"
	_ "github.com/golang-migrate/migrate/v3/database/postgres"
	_ "github.com/golang-migrate/migrate/v3/source/file"
)

const (
	InfoDbClosed = "Storage is closed. App is shuting down"
)

// @title Test-task
// @version 1.0
// @description Service for test task Effective mobile

// @host localhost:8080
// @BasePath /api/v1/
func main() {

	ctx := context.Background()

	cfg := config.MustRead()

	log := logger.New(cfg.Log)

	// connect to db and start migrations
	storage, err := postgres.New(ctx, log, cfg.DbConnString)
	if err != nil {
		log.Error("can't connect to storage", "err", err.Error())

		os.Exit(1)
	}

	err = startMigrations(log, cfg.DbConnString)
	if err != nil {
		log.Error("can't start migrations", "err", err)

		os.Exit(1)
	}

	// init api with services
	api := api.New(log, storage)

	srv := http.Server{
		Addr:    cfg.ServerHost + ":" + cfg.ServerPort,
		Handler: api.Router,
	}

	//graceful shutdown
	chanErrors := make(chan error, 1)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("Server started", "Addres", srv.Addr)
		chanErrors <- srv.ListenAndServe()
	}()

	go func() {
		log.Info("Started to ping databse")
		for {
			time.Sleep(5 * time.Second)
			err := storage.Ping(ctx)
			if err != nil {
				chanErrors <- err
				break
			}
		}

	}()

	log.Info("http server is runned", "addres", srv.Addr)

	log.Info("App is started")

	select {
	case err := <-chanErrors:
		log.Error("Shutting down. Critical error:", "err", err)

		shutdown <- syscall.SIGTERM
	case sig := <-shutdown:
		log.Error("received signal, starting graceful shutdown", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("server graceful shutdown failed", "err", err)
			err = srv.Close()
			if err != nil {
				log.Error("forced shutdown failed", "err", err)
			}
		}

		storage.Close()

		log.Info(InfoDbClosed)

		log.Info("shutdown completed")

	}

}

func startMigrations(log *slog.Logger, connString string) error {
	m, err := migrate.New("file://migrations", connString) // DEBUG: ../../migrations"
	if err != nil {
		return fmt.Errorf("can't start migration driver:%w", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Warn("Migarate didn't run. Nothing to change")
			return nil
		}
		return fmt.Errorf("failed to do migrations:%w", err)

	}

	return nil
}
