package xprometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/Deimvis-go/xprometheus/xprometheus/xprommetric"
)

type MetricGroup interface {
	Collectors() []prometheus.Collector
}

func NewCounterHistGroup(c *prometheus.CounterVec, h *prometheus.HistogramVec) *CounterHistMetricGroup {
	return &CounterHistMetricGroup{
		counter: c,
		hist:    h,
	}
}

type CounterHistMetricGroup struct {
	counter *prometheus.CounterVec
	hist    *prometheus.HistogramVec
}

func (chg *CounterHistMetricGroup) Collectors() []prometheus.Collector {
	return []prometheus.Collector{chg.counter, chg.hist}
}

func (chg *CounterHistMetricGroup) C() *prometheus.CounterVec {
	return chg.counter
}

func (chg *CounterHistMetricGroup) H() *prometheus.HistogramVec {
	return chg.hist
}

func NewIntervalGroup(
	startC *prometheus.CounterVec,
	finishC *prometheus.CounterVec,
	durationH DurationHistogramVec,
) IntervalMetricGroup {
	durationOrigScaler := durationH.Scaler()
	mustImplement[xprommetric.SafeUnitScaler[time.Duration, float64]](durationOrigScaler)
	durationScaler := durationOrigScaler.(xprommetric.SafeUnitScaler[time.Duration, float64])
	return IntervalMetricGroup{
		startC:         startC,
		finishC:        finishC,
		durationH:      durationH,
		durationScaler: durationScaler,
	}
}

type IntervalMetricGroup struct {
	startC         *prometheus.CounterVec
	finishC        *prometheus.CounterVec
	durationH      DurationHistogramVec
	durationScaler xprommetric.SafeUnitScaler[time.Duration, float64]
}

func (ig IntervalMetricGroup) Collectors() []prometheus.Collector {
	return []prometheus.Collector{ig.startC, ig.finishC, ig.durationH.Metric()}
}

func (ig IntervalMetricGroup) StartC() *prometheus.CounterVec {
	return ig.startC
}

func (ig IntervalMetricGroup) FinishC() *prometheus.CounterVec {
	return ig.finishC
}

func (ig IntervalMetricGroup) DurationH() DurationHistogramVec {
	return ig.durationH
}

func (ig IntervalMetricGroup) ScaleDuration(d time.Duration) float64 {
	return ig.durationScaler.SafeScale(d)
}

func (ig IntervalMetricGroup) Record(ls prometheus.Labels, fn RecordFn) {
	Record(func(rc RecordControl) {
		rc.SetLabels(ls)
		start := time.Now()
		ig.startC.With(rc.GetLabels()).Inc()

		fn(rc)
		elapsed := ig.durationScaler.SafeScale(time.Since(start))

		ig.finishC.With(rc.GetLabels()).Inc()
		ig.durationH.Metric().With(rc.GetLabels()).Observe(elapsed)
	})
}
