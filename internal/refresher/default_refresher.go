package refresher

import (
	"context"
	"log/slog"
	"time"

	"github.com/kodehat/netcupscp-exporter/internal/authenticator"
	"github.com/kodehat/netcupscp-exporter/internal/metrics"
)

type DefaultRefresher struct {
	authenticator   authenticator.Authenticator
	metricsUpdater  metrics.MetricsUpdater
	refreshInterval time.Duration
}

var _ Refresher = DefaultRefresher{}

// authExpiryBuffer is the duration before actual expiry when we consider the token to be expired.
const authExpiryBuffer time.Duration = 15 * time.Second

func NewDefaultRefresher(authenticator authenticator.Authenticator, metricsUpdater metrics.MetricsUpdater, refreshInterval time.Duration) DefaultRefresher {
	return DefaultRefresher{
		authenticator:   authenticator,
		metricsUpdater:  metricsUpdater,
		refreshInterval: refreshInterval,
	}
}

func (dr DefaultRefresher) refresh(ctx context.Context) (isAuthError bool, isMetricError bool, err error) {
	now := time.Now()
	if dr.authenticator.IsAuthenticationExpired(now.Add(authExpiryBuffer)) {
		slog.Debug("access token near expiration, refreshing authentication", "expiry", dr.authenticator.GetAuthData().Expiry)
		_, err := dr.authenticator.Authenticate(ctx)
		if err != nil {
			slog.Error("error refreshing authentication", "error", err)
			return true, false, err
		}
	}
	err = dr.metricsUpdater.UpdateMetrics(ctx)
	if err != nil {
		slog.Error("error while updating metrics", "error", err)
		return false, true, err
	}
	slog.Debug("metrics have been updated successfully")
	return false, false, nil
}

func (dr DefaultRefresher) StartRefreshMetricsPeriodically(ctx context.Context) {
	ticker := time.NewTicker(dr.refreshInterval)
	defer ticker.Stop()
	slog.Info("starting periodic metrics update", "interval", dr.refreshInterval.String())
	dr.refresh(ctx) // Run once immediately.
	for {
		select {
		case <-ticker.C:
			isAuthError, isMetricError, err := dr.refresh(ctx)
			if err != nil {
				if isAuthError {
					slog.Error("authentication error occurred during metrics refresh, will exit now", "error", err)
					return
				} else if isMetricError {
					slog.Warn("metrics update error occurred during metrics refresh", "error", err)
				} else {
					slog.Warn("unknown error occurred during metrics refresh", "error", err)
				}
			}
		case <-ctx.Done():
			slog.Debug("stopping updating metrics")
			return
		}
	}
}
