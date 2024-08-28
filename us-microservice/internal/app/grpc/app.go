package grpcapp

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"urlSh/internal/config"
	"urlSh/internal/domain/models"
	"urlSh/internal/grpc/server"
	"urlSh/internal/kafka"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type App struct {
	log        *slog.Logger
	config     *config.Config
	gRPCServer *grpc.Server
	kafkaCh    chan models.Url
}

// New creates new gRPC server app.
func New(
	log *slog.Logger,
	config *config.Config,
	urlService server.URLShortener,
	kafkaCh chan models.Url,
) *App {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, logging.PayloadSent,
		),
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("Recovered from panic", slog.Any("panic", p))
			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...),
	), grpc.ConnectionTimeout(config.Grpc.Timeout))

	server.Register(gRPCServer, urlService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		config:     config,
		kafkaCh:    kafkaCh,
	}
}

// InterceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

// Run runs gRPC server.
func (a *App) Run(ctx context.Context) {
	const op = "grpcapp.Run"

	wg := &sync.WaitGroup{}

	wg.Add(1)
	producer, err := kafka.NewProducer(a.log, a.config, a.kafkaCh)
	if err != nil {
		a.log.Error("%s: %w", op, err)
		ctx.Done()
		return
	}
	go producer.RunProducing(ctx, wg)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.config.Grpc.Port))
	if err != nil {
		a.log.Error("%s: %w", op, err)
	}

	a.log.Info("grpc server started", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		a.log.Error("%s: %w", op, err)
	}
	wg.Wait()

}

// Stop stops gRPC server.
func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.config.Grpc.Port))

	a.gRPCServer.GracefulStop()
}
