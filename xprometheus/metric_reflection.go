package xprometheus

import "github.com/prometheus/client_golang/prometheus"

type MetricReflection[T prometheus.Collector] struct {
}
