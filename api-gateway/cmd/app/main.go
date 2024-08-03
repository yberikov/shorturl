package main

import (
	"apiGW/internal/app"
	"apiGW/internal/config"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	application := app.New(log, cfg)

	log.Info("starting api-gateway", slog.String("port", cfg.Port))

	ctx, cancel := context.WithCancel(context.Background())

	go application.Run()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	log.Info("stopping server: releasing all resources")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, time.Second*5)
	defer shutdownCancel()
	if err := application.Stop(shutdownCtx); err != nil {
		log.Error("failed to gracefully stop server", slog.String("error", err.Error()))
	}

	log.Info("server gracefully stopped")
}
