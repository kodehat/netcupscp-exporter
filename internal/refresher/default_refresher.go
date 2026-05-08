package refresher

import (
	"context"
	"log/slog"
	"time"

	"github.com/kodehat/netcupscp-exporter/internal/metrics"
)

type DefaultRefresher struct {
	metricsUpdater  metrics.MetricsUpdater
	refreshInterval time.Duration
}

var _ Refresher = DefaultRefresher{}

func NewDefaultRefresher(metricsUpdater metrics.MetricsUpdater, refreshInterval time.Duration) DefaultRefresher {
	return DefaultRefresher{
		metricsUpdater:  metricsUpdater,
		refreshInterval: refreshInterval,
	}
}

func (dr DefaultRefresher) refresh(ctx context.Context) error {
	err := dr.metricsUpdater.UpdateMetrics(ctx)
	if err != nil {
		slog.Error("error while updating metrics", "error", err)
		return err
	}
	slog.Debug("metrics have been updated successfully", "next_update", time.Now().Add(dr.refreshInterval))
	return nil
}

func (dr DefaultRefresher) StartRefreshMetricsPeriodically(ctx context.Context) {
	ticker := time.NewTicker(dr.refreshInterval)
	defer ticker.Stop()
	slog.Info("starting periodic metrics update", "interval", dr.refreshInterval.String())
	dr.refresh(ctx) // Run once immediately.
	for {
		select {
		case <-ticker.C:
			if err := dr.refresh(ctx); err != nil {
				slog.Warn("metrics update error occurred during metrics refresh", "error", err)
			}
		case <-ctx.Done():
			slog.Debug("stopping updating metrics")
			return
		}
	}
}
