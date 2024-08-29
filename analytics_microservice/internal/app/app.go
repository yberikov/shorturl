package app

import (
	grpcapp "analys/internal/app/grpc"
	"analys/internal/config"
	"analys/internal/services"
	"analys/internal/storage/clickhouse"
	"log/slog"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, cfg *config.Config) *App {
	storage, err := clickhouse.New(cfg.Storage)
	if err != nil {
		if err != nil {
			panic(err)
		}
	}

	analyticsService := services.New(log, storage)

	grpcApp := grpcapp.New(log, cfg, analyticsService, analyticsService)

	return &App{
		GRPCServer: grpcApp,
	}
}
