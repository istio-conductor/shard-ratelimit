package config

import (
	"github.com/istio-conductor/shard-ratelimit/prom"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

// Metrics for an individual rate limit config entry.
type Metrics struct {
	TotalHits               prometheus.Counter
	OverLimit               prometheus.Counter
	NearLimit               prometheus.Counter
	OverLimitWithLocalCache prometheus.Counter
	WithinLimit             prometheus.Counter
}

// NewMetrics Create a new rate limit Metrics for a config entry.
func NewMetrics(kv string) Metrics {
	var domain = ""
	splited := strings.SplitN(kv, ".", 2)
	if len(splited) == 2 {
		domain = splited[0]
		kv = splited[1]
	}
	m := Metrics{}
	m.TotalHits = prom.TotalHits.WithLabelValues(domain, kv)
	m.OverLimit = prom.OverLimit.WithLabelValues(domain, kv)
	m.NearLimit = prom.NearLimit.WithLabelValues(domain, kv)
	m.WithinLimit = prom.WithinLimit.WithLabelValues(domain, kv)
	return m
}
