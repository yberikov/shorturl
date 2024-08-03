package clientConn

import (
	"apiGW/internal/config"
	an "gitea.com/yberikov/us-protos/gen/analytics-service"
	au "gitea.com/yberikov/us-protos/gen/auth-service"
	us "gitea.com/yberikov/us-protos/gen/us-service"
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
