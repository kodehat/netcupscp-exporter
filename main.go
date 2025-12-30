package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/kodehat/netcupscp-exporter/internal/authenticator"
	"github.com/kodehat/netcupscp-exporter/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	refreshToken := os.Getenv("NETCUP_REFRESH_TOKEN")
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	if host == "" {
		log.Println("No host provided, using default (bind to all interfaces)")
	}
	if port == "" {
		log.Println("No port provided, using default (2008)")
		port = "2008"
	}
	authenticator := authenticator.NewDefaultAuthenticator("", refreshToken)
	authData, err := authenticator.Authenticate(context.TODO())
	if err != nil {
		log.Panicf("Error during authentication: %v", err)
	}
	metricsUpdater := metrics.NewMetricsUpdater(authData)
	go metricsUpdater.UpdateMetricsPeriodically(context.Background(), 10*time.Second)

	listenAddr := net.JoinHostPort(host, port)
	fmt.Printf("Listening on %q\n", listenAddr)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(listenAddr, nil)
}
