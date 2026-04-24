package xprometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/Deimvis-go/xprometheus/xprometheus/xprommetric"
)

type DurationHistogram = xprommetric.HavingUnit[
	prometheus.Histogram,
	time.Duration,
	float64,
]
type DurationHistogramVec = xprommetric.HavingUnit[
	*prometheus.HistogramVec,
	time.Duration,
	float64,
]
