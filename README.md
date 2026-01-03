# netcupscp-exporter

A small Prometheus exporter that exports metrics about servers from Netcup SCP (server control panel).

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Metrics](#metrics)
- [Docker](#docker)
- [Development](#development)
- [License](#license)

## Features

- Exposes Prometheus metrics for Netcup SCP resources
- Periodically refreshes data and updates metrics
- Automatic refresh of used access token
- Lightweight Go binary with optional Docker image

## Installation

Build the binary with Go (Go 1.25+ recommended):

```sh
go mod download
go build -o netcupscp-exporter ./
```

## Usage

Run the exporter with flags or environment variables. For example:

```sh
# Using flags
./netcupscp-exporter --port=2008 --refresh-token=<token> --log-level=debug

# Or using environment variables
PORT=2008 REFRESH_TOKEN=<token> ./netcupscp-exporter
```

See the available flags in the source (`main.go` and the `flags` package).

## Configuration

Configuration can be provided via environment variables or command-line flags. Flags override environment variables when both are provided.

Environment variables (defaults shown):

- `HOST` — host to bind the HTTP server to (default: all interfaces / empty)
- `PORT` — port to bind the HTTP server to (default: `2008`)
- `REFRESH_TOKEN` — Netcup SCP refresh token (default: empty)
- `REVOKE_TOKEN` — set to `true` to revoke the given refresh token and exit (default: `false`)
- `GET_TOKEN_DETAILS` — set to `true` to print details about the refresh token and exit (default: `false`)
- `LOG_LEVEL` — logging level (default: `info`; options: `debug`, `info`, `warn`, `error`)
- `LOG_JSON` — set to `true` to enable JSON formatted logging (default: `false`)

Command-line flags (override env vars):

- `--host` string (bind host)
- `--port` string (bind port)
- `--refresh-token` string (Netcup SCP refresh token)
- `--revoke-token` bool (revoke refresh token and exit)
- `--get-token-details` bool (print token details and exit)
- `--log-level` string (logging level)
- `--log-json` bool (enable JSON logging)

## Metrics

The exporter exposes Prometheus metrics (HTTP) on the configured address and path (commonly `/metrics`). Configure Prometheus to scrape the exporter endpoint.

**Collected metrics** (prometheus names prefixed with `ncscp_`):

- **ncscp_build_info**: gauge — constant `1` labeled by `goversion`, `revision`, `version`.
- **ncscp_cpu_cores**: gauge — number of CPU cores; labels: `servername`, `servernickname`.
- **ncscp_memory_bytes**: gauge — amount of memory in bytes; labels: `servername`, `servernickname`.
- **ncscp_monthlytraffic_in_bytes**: gauge — monthly incoming traffic in bytes; labels: `servername`, `servernickname`, `month`, `year`, `mac`.
- **ncscp_monthlytraffic_out_bytes**: gauge — monthly outgoing traffic in bytes; labels: `servername`, `servernickname`, `month`, `year`, `mac`.
- **ncscp_monthlytraffic_total_bytes**: gauge — total monthly traffic in bytes; labels: `servername`, `servernickname`, `month`, `year`, `mac`.
- **ncscp_server_start_time_seconds**: gauge — server start time (seconds since epoch); labels: `servername`, `servernickname`.
- **ncscp_ip_info**: gauge — IP addresses assigned to a server; labels: `servername`, `servernickname`, `mac`, `ip`, `type`.
- **ncscp_interface_throttled**: gauge — interface throttled (1) or not (0); labels: `servername`, `servernickname`, `mac`, `status`.
- **ncscp_server_status**: gauge — online (1) / offline (0); labels: `servername`, `servernickname`, `status`.
- **ncscp_rescue_active**: gauge — rescue system active (1) / inactive (0); labels: `servername`, `servernickname`, `status`.
- **ncscp_reboot_recommended**: gauge — reboot recommended (1) / not (0); labels: `servername`, `servernickname`, `status`.
- **ncscp_disk_capacity_bytes**: gauge — available storage space in bytes; labels: `servername`, `servernickname`, `driver`, `name`.
- **ncscp_disk_used_bytes**: gauge — used storage space in bytes; labels: `servername`, `servernickname`, `driver`, `name`.
- **ncscp_disk_optimization**: gauge — optimization recommended (1) / not (0); labels: `servername`, `servernickname`, `status`.

## Docker

You can build a Docker image using the provided `Dockerfile` and run it passing environment variables or flags:

```sh
docker build -t codehat/netcupscp-exporter:local .
docker run --rm -e REFRESH_TOKEN=<token> -p 2008:2008 codehat/netcupscp-exporter:local
```

Adjust ports and environment variables to your environment.

## Development

- Install dependencies: `go mod download`
- Run the exporter locally: `go run ./...`

You can regenerate Netcup SCP client code with `client/generate.go` file.

## License

[MIT](https://www.tldrlegal.com/license/mit-license)