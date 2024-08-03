package grpcapp

import (
	"analys/internal/config"
	"analys/internal/grpc/server"
	"analys/internal/kafka"
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"log/slog"
	"net"
	"sync"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type App struct {
	log              *slog.Logger
	config           *config.Config
	gRPCServer       *grpc.Server
	saveStatsService kafka.SaveStatsService
}

func New(
	log *slog.Logger,
	config *config.Config,
	getStatsService server.GetStatsService,
	saveStatsService kafka.SaveStatsService,
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

	server.Register(gRPCServer, getStatsService)

	return &App{
		log:              log,
		gRPCServer:       gRPCServer,
		config:           config,
		saveStatsService: saveStatsService,
	}
}

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

// Run runs gRPC server.
func (a *App) Run(ctx context.Context) {
	wg := &sync.WaitGroup{}

	a.log.Info("kafka consumer starting")

	wg.Add(1)
	consumer, err := kafka.RunConsumer(ctx, wg, a.log, a.config, a.saveStatsService)
	if err != nil {
		panic(err)
	}
	defer func(consumer sarama.ConsumerGroup) {
		err := consumer.Close()
		if err != nil {
			a.log.Error("error on closing consumer:", slog.String("err", err.Error()))
		}
	}(consumer)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.config.Grpc.Port))
	if err != nil {
		a.log.Error("error on listening tcp:", slog.String("err", err.Error()))
	}

	a.log.Info("grpc server started", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		a.log.Error("error on listening tcp:", slog.String("err", err.Error()))
	}

	wg.Wait()
}

// Stop stops gRPC server.
func (a *App) Stop() {
	a.log.Info("stopping gRPC server", slog.Int("port", a.config.Grpc.Port))

	a.gRPCServer.GracefulStop()
}
