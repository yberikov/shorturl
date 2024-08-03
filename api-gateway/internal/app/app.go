package app

import (
	"apiGW/internal/config"
	clientConn "apiGW/internal/http-server/client"
	"apiGW/internal/http-server/server"
	"context"
	"log/slog"
	"net/http"
)

type App struct {
	log    *slog.Logger
	cfg    *config.Config
	server *http.Server
}

func New(log *slog.Logger, cfg *config.Config) *App {

	client, err := clientConn.New(cfg, log)
	if err != nil {
		panic(err)
	}

	return &App{
		log:    log,
		cfg:    cfg,
		server: server.New(cfg, client),
	}
}

func (a *App) Run() {
	go func() {
		if err := a.server.ListenAndServe(); err != nil {
			a.log.Info("serverRun: shutdown or closed")
		}
	}()
}

func (a *App) Stop(ctx context.Context) error {
	ctx.Done()
	err := a.server.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}
