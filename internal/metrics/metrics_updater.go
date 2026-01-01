package metrics

import "context"

type MetricsUpdater interface {
	UpdateMetrics(context.Context) error
}
