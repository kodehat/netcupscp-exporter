package metrics

import "github.com/prometheus/client_golang/prometheus"

// mergeLabels merges two sets of Prometheus labels.
func mergeLabels(base prometheus.Labels, extra prometheus.Labels) prometheus.Labels {
	merged := prometheus.Labels{}
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range extra {
		merged[k] = v
	}
	return merged
}
