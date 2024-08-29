package clientConn

import (
	"apiGW/internal/config"
	an "github.com/yberikov/us-protos/gen/analytics-microservice"
	au "github.com/yberikov/us-protos/gen/auth-microservice"
	us "github.com/yberikov/us-protos/gen/us-microservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
)

type ClientConn struct {
	Log                *slog.Logger
	UrlShortenerClient us.UrlShorteningServiceClient
	AuthClient         au.AuthServiceClient
	AnalyticsClient    an.AnalyticsServiceClient
}

func New(cfg *config.Config, log *slog.Logger) (*ClientConn, error) {

	usConn, err := grpc.NewClient(cfg.UrlAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	authConn, err := grpc.NewClient(cfg.AuthAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	anConn, err := grpc.NewClient(cfg.AnAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &ClientConn{
		UrlShortenerClient: us.NewUrlShorteningServiceClient(usConn),
		AuthClient:         au.NewAuthServiceClient(authConn),
		AnalyticsClient:    an.NewAnalyticsServiceClient(anConn),
		Log:                log,
	}, nil
}
