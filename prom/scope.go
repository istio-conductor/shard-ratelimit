package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "ratelimit"
const (
	ComponentService = "service"
	ComponentRedis   = "redis"
)

var ResponseTime = promauto.NewHistogram(prometheus.HistogramOpts{
	Namespace: namespace,
	Subsystem: ComponentService,
	Name:      "response_time_milliseconds",
	Buckets: []float64{
		0.1, 0.5, 1, 3, 5, 10, 25, 50, 100, 250, 500, 1000,
	},
})

var Requests = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: ComponentService,
	Name:      "total_requests",
	ConstLabels: map[string]string{
		"grpc_method": "ShouldRateLimit",
	},
})

var RateLimitError = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: ComponentService,
	Name:      "should_rate_limit_error",
}, []string{"err_type"})

var RedisError = RateLimitError.WithLabelValues("redis_error")
var ServiceError = RateLimitError.WithLabelValues("service_error")

var ConfigLoadSuccess = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: ComponentService,
	Name:      "config_load_success",
})

var ConfigLoadError = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: ComponentService,
	Name:      "config_load_error",
})

var TotalHits = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: ComponentService,
	Name:      "rate_limit_total_hits",
}, []string{
	"domain", "kv",
})

var OverLimit = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: ComponentService,
	Name:      "rate_limit_over_limit",
}, []string{
	"domain", "kv",
})

var NearLimit = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: ComponentService,
	Name:      "rate_limit_near_limit",
}, []string{
	"domain", "kv",
})

var WithinLimit = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: ComponentService,
	Name:      "rate_limit_within_limit",
}, []string{
	"domain", "kv",
})

type PoolStat struct {
	Active prometheus.Gauge
	Total  prometheus.Counter
	Close  prometheus.Counter
}

func NewPoolStat(prefix string) PoolStat {
	return PoolStat{
		promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: ComponentRedis,
			Name:      prefix + "_active",
		}),
		promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: ComponentRedis,
			Name:      prefix + "_total",
		}),
		promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: ComponentRedis,
			Name:      prefix + "_close",
		}),
	}
}

var RedisPerSecondPool = NewPoolStat("per_second_pool")
var RedisPool = NewPoolStat("pool")
