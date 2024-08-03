package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"urlSh/internal/app"
	"urlSh/internal/config"
)

func main() {
	cfg := config.MustLoad()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	application := app.New(log, cfg)

	ctx, cancel := context.WithCancel(context.Background())

	go application.GRPCServer.Run(ctx)

	// Graceful shutdown

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	<-done
	cancel()

	application.GRPCServer.Stop()
	log.Info("Gracefully stopped")
}
