package flags

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
)

const (
	envHost         = "HOST"
	envPort         = "PORT"
	envRefreshToken = "REFRESH_TOKEN"
	envLogLevel     = "LOG_LEVEL"
	envLogJson      = "LOG_JSON"
)

type Flags struct {
	Host         string
	Port         string
	RefreshToken string
	logLevel     string
	logJson      bool
}

var F Flags

func Load() {
	// Set up default values from environment variables.
	host := getenvOrDefault(envHost, "")
	port := getenvOrDefault(envPort, "2008")
	refreshToken := getenvOrDefault(envRefreshToken, "")
	logLevel := getenvOrDefault(envLogLevel, "info")
	logJson := false
	if getenvOrDefault(envLogJson, "false") == "true" {
		logJson = true
	}

	// Allow overriding via command-line flags.
	flags := Flags{}
	flag.StringVar(&flags.Host, "host", host, "Host to bind the HTTP server to (default: all interfaces)")
	flag.StringVar(&flags.Port, "port", port, "Port to bind the HTTP server to (default: 2008)")
	flag.StringVar(&flags.RefreshToken, "refresh-token", refreshToken, "Netcup SCP refresh token for authentication. Can be ommitted for first time setup.")
	flag.StringVar(&flags.logLevel, "log-level", logLevel, "Logging level (debug, info, warn, error)")
	flag.BoolVar(&flags.logJson, "log-json", logJson, "Enable JSON formatted logging")
	flag.Parse()

	// Store the final flag values.
	F = flags
}

func (f Flags) GetLogLevel() (slog.Level, error) {
	var level slog.Level
	err := level.UnmarshalText([]byte(f.logLevel))
	return level, err
}

func (f Flags) GetLogHandler(w io.Writer) slog.Handler {
	logLevel, err := f.GetLogLevel()
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal log level: %w\n", err))
	}
	logHandlerOptions := &slog.HandlerOptions{Level: logLevel}
	if f.logJson {
		return slog.NewJSONHandler(w, logHandlerOptions)
	}
	return slog.NewTextHandler(w, logHandlerOptions)
}
