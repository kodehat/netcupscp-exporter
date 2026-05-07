package collector

import (
	"context"
)

type ApiInfo struct {
	statusCode int
}

type ApiCollector interface {
	CollectApiData(context context.Context) (ApiInfo, error)
}
