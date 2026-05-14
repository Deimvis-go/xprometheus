package prom

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/Deimvis-go/xprometheus/prom/prommetric"
)

type DurationHistogram = prommetric.HavingUnit[
	prometheus.Histogram,
	time.Duration,
	float64,
]
type DurationHistogramVec = prommetric.HavingUnit[
	*prometheus.HistogramVec,
	time.Duration,
	float64,
]
