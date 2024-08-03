package server

import (
	"analys/internal/storage"
	"context"
	"errors"
	an "gitea.com/yberikov/us-protos/gen/analytics-service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetStatsService interface {
	GetURLStats(context.Context, string) (int64, error)
	LogURLAccess(context.Context, string, int64) (bool, error)
}

type serverAPI struct {
	an.UnimplementedAnalyticsServiceServer
	analyticsService GetStatsService
}

func Register(gRPCServer *grpc.Server, analyticsService GetStatsService) {
	an.RegisterAnalyticsServiceServer(gRPCServer, &serverAPI{analyticsService: analyticsService})
}

func (s *serverAPI) GetURLStats(
	ctx context.Context,
	in *an.GetURLStatsRequest,
) (*an.GetURLStatsResponse, error) {
	if in.Url == "0" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	stats, err := s.analyticsService.GetURLStats(ctx, in.Url)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get stats")
	}

	return &an.GetURLStatsResponse{
		TotalAccesses: stats,
	}, nil
}

func (s *serverAPI) LogURLAccess(
	ctx context.Context,
	in *an.LogURLAccessRequest,
) (*an.LogURLAccessResponse, error) {
	if in.Url == "0" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	success, err := s.analyticsService.LogURLAccess(ctx, in.Url, in.UserId)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get access")
	}

	return &an.LogURLAccessResponse{
		Success: success,
	}, nil
}
