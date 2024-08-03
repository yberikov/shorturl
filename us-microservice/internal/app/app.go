package app

import (
	"log/slog"
	grpcapp "urlSh/internal/app/grpc"
	"urlSh/internal/config"
	"urlSh/internal/domain/models"
	"urlSh/internal/services"
	"urlSh/internal/storage/mongodb"
	"urlSh/internal/storage/redis"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, cfg *config.Config) *App {
	storage, err := mongodb.New(cfg.Storage.Path, cfg.Storage.Database, cfg.Storage.Collection)
	if err != nil {
		panic(err)
	}
	cache, err := redis.New(cfg.CachePath)
	if err != nil {
		panic(err)
	}
	kafkaCh := make(chan models.Url)

	urlService := services.New(log, storage, cache, cfg.Ttl, kafkaCh)

	grpcApp := grpcapp.New(log, cfg, urlService, kafkaCh)

	return &App{
		GRPCServer: grpcApp,
	}
}
