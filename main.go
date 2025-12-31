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

func run(ctx context.Context, flags flags.Flags, _ io.Reader, stdout, stderr io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := slog.New(flags.GetLogHandler(stdout))
	slog.SetDefault(logger)

	authenticator := authenticator.NewDefaultAuthenticator("", flags.RefreshToken)
	authData, err := authenticator.Authenticate(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "error during authentication: %s\n", err)
		return err
	}

	registry := metrics.Load()

	metricsUpdater := metrics.NewMetricsUpdater(authData)
	go metricsUpdater.UpdateMetricsPeriodically(ctx, metricRefreshInterval)

	// Create http server for Prometheus metrics.
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(flags.Host, flags.Port),
		Handler: promhttp.HandlerFor(registry, promhttp.HandlerOpts{EnableOpenMetrics: true}),
	}

	go func() {
		logger.Info("server is now accepting connections", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(stderr, "error listening and serving: %s\n", err)
		}
	}()
	var wg sync.WaitGroup
	wg.Go(func() {
		<-ctx.Done()
		// Make a new context for the shutdown.
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(stderr, "error shutting down http server: %s\n", err)
		}
	})
	wg.Wait()
	return nil
}
