package services

import (
	"context"
	"encoding/json"
	"log/slog"

	"analys/internal/domain/models"
)

type StatsStorage interface {
	SaveStats(ctx context.Context, userID int64, urlText string) error
	GetURLStats(ctx context.Context, url string) (int64, error)
	LogURLAccess(ctx context.Context, url string, userId int64) (bool, error)
}

type AnalyticsService struct {
	log        *slog.Logger
	statsStore StatsStorage
}

func New(log *slog.Logger, statsStore StatsStorage) *AnalyticsService {
	return &AnalyticsService{
		log:        log,
		statsStore: statsStore,
	}
}

func (s *AnalyticsService) SaveRecord(ctx context.Context, content []byte) error {
	var msg models.UrlStat
	err := json.Unmarshal(content, &msg)
	if err != nil {
		return err
	}
	err = s.statsStore.SaveStats(ctx, msg.UserId, msg.UrlText)
	if err != nil {
		s.log.Error("failed to save stats", slog.String("err", err.Error()))
		return err
	}

	return nil
}

func (s *AnalyticsService) GetURLStats(ctx context.Context, url string) (int64, error) {

	stats, err := s.statsStore.GetURLStats(ctx, url)
	if err != nil {
		s.log.Error("failed to get total stats", slog.String("err", err.Error()))
		return 0, err
	}

	return stats, nil
}

func (s *AnalyticsService) LogURLAccess(ctx context.Context, url string, userId int64) (bool, error) {
	success, err := s.statsStore.LogURLAccess(ctx, url, userId)
	if err != nil {
		s.log.Error("failed to get total stats", slog.String("err", err.Error()))
		return false, err
	}

	return success, nil
}
