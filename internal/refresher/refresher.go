package refresher

import "context"

type Refresher interface {
	StartRefreshMetricsPeriodically(context.Context)
}
