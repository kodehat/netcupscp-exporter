package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/kodehat/netcupscp-exporter/internal/authenticator"
	"github.com/kodehat/netcupscp-exporter/internal/collector"
	"github.com/kodehat/netcupscp-exporter/internal/flags"
	"github.com/kodehat/netcupscp-exporter/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var metricRefreshInterval = 30 * time.Second

func main() {
	ctx := context.Background()
	flags.Load()

	if err := run(ctx, flags.F, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, flags flags.Flags, _ io.Reader, stdout, _ io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := slog.New(flags.GetLogHandler(stdout))
	slog.SetDefault(logger)

	authenticator := authenticator.NewDefaultAuthenticator("", flags.RefreshToken)
	if flags.RevokeToken {
		logger.Info("revoking refresh token as requested")
		err := authenticator.Revoke(ctx)
		if err != nil {
			logger.Error("error revoking refresh token", "error", err)
			return err
		}
		logger.Info("successfully revoked refresh token, the application will now exit")
		return nil
	}
	authResult, err := authenticator.Authenticate(ctx)
	if err != nil {
		logger.Error("error during authentication", "error", err)
		return err
	}
	if authResult.IsNewDevice {
		logger.Warn("first-time setup: obtained new refresh token, please store it for future use", "refresh_token", authResult.AuthData.RefreshToken)
		logger.Info("the application will now exit, please restart it with the new refresh token")
		return nil
	}

	// Getting token details requires a valid access token.
	if flags.GetTokenDetails {
		logger.Info("getting token details as requested")
		userInfo, err := authenticator.GetUserInfo(ctx)
		if err != nil {
			logger.Error("error getting token details", "error", err)
			return err
		}
		logger.Info("token details obtained successfully", "name", userInfo.Name, "email", userInfo.Email, "username", userInfo.PreferredUsername)
		return nil
	}

	registry := metrics.Load()

	serverCollector := collector.NewDefaultServerCollector(authResult.AuthData)
	metricsUpdater := metrics.NewDefaultMetricsUpdater(serverCollector)

	go metricsUpdater.UpdateMetricsPeriodically(ctx, metricRefreshInterval)

	// Create http server for Prometheus metrics.
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{EnableOpenMetrics: true}))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(flags.Host, flags.Port),
		Handler: mux,
	}

	go func() {
		logger.Info("server is now accepting connections", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("error listening and serving", "error", err)
		}
	}()
	var wg sync.WaitGroup
	wg.Go(func() {
		<-ctx.Done()
		// Use a new context for shutdown.
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("error shutting down http server", "error", err)
		}
	})
	wg.Wait()
	return nil
}
