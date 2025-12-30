package collector

import (
	"context"

	"github.com/kodehat/netcupscp-exporter/internal/client"
)

type ServerInfo struct {
	*client.Server
}

type ServerCollector interface {
	CollectServerData(context context.Context) ([]ServerInfo, error)
}
