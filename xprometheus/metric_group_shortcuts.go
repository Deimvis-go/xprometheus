package xprometheus

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/Deimvis-go/xprometheus/xprometheus/xprommetric"
)

type RPSLatencyGroupOption func(*rpsLatencyGroupCfg)

func NewRPSLatencyGroup(
	baseName string,
	startLabels prometheus.ConstrainableLabels,
	finishLabels prometheus.ConstrainableLabels,
	opts ...RPSLatencyGroupOption,
) *RPSLatencyGroup {
	cfg := defaultRPSLatencyGroupConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	g := &RPSLatencyGroup{
		IntervalMetricGroup: NewIntervalGroup(
			prometheus.V2.NewCounterVec(prometheus.CounterVecOpts{
				CounterOpts: prometheus.CounterOpts{
					Name: baseName + cfg.naming.startSuffix + cfg.naming.counterSuffix,
				},
				VariableLabels: startLabels,
			}),
			prometheus.V2.NewCounterVec(prometheus.CounterVecOpts{
				CounterOpts: prometheus.CounterOpts{
					Name: baseName + cfg.naming.finishSuffix + cfg.naming.counterSuffix,
				},
				VariableLabels: finishLabels,
			}),
			xprommetric.NewHavingUnit(
				prometheus.V2.NewHistogramVec(prometheus.HistogramVecOpts{
					HistogramOpts: prometheus.HistogramOpts{
						Name:    baseName + cfg.naming.durationSuffix + cfg.naming.histogramSuffix,
						Buckets: cfg.latencyBuckets,
					},
					VariableLabels: finishLabels,
				}),
				xprommetric.DurationToSecondsScaler,
			),
		),
	}
	return g
}

func WithStartSuffix(s string) RPSLatencyGroupOption {
	return func(cfg *rpsLatencyGroupCfg) {
		cfg.naming.startSuffix = s
	}
}

func WithFinishSuffix(s string) RPSLatencyGroupOption {
	return func(cfg *rpsLatencyGroupCfg) {
		cfg.naming.finishSuffix = s
	}
}

func WithDurationSuffix(s string) RPSLatencyGroupOption {
	return func(cfg *rpsLatencyGroupCfg) {
		cfg.naming.durationSuffix = s
	}
}

func WithCounterSuffix(s string) RPSLatencyGroupOption {
	return func(cfg *rpsLatencyGroupCfg) {
		cfg.naming.counterSuffix = s
	}
}

func WithHistogramSuffix(s string) RPSLatencyGroupOption {
	return func(cfg *rpsLatencyGroupCfg) {
		cfg.naming.histogramSuffix = s
	}
}

func WithLatencyBuckets(vs []float64) RPSLatencyGroupOption {
	return func(cfg *rpsLatencyGroupCfg) {
		cfg.latencyBuckets = vs
	}
}

type RPSLatencyGroup struct {
	IntervalMetricGroup
}

type rpsLatencyGroupCfg struct {
	naming         intervalMetricsNaming
	latencyBuckets []float64
}

type intervalMetricsNaming struct {
	// naming: {base_name}{start/finish/duration suffix}{counter/hist suffix}
	startSuffix     string
	finishSuffix    string
	durationSuffix  string
	counterSuffix   string
	histogramSuffix string
}

var defaultRPSLatencyGroupConfig = rpsLatencyGroupCfg{
	naming: intervalMetricsNaming{
		startSuffix:     "_start",
		finishSuffix:    "_finish",
		durationSuffix:  "_duration",
		counterSuffix:   "_count",
		histogramSuffix: "", // NOTE: it's a good practice to leave a suffix with unit (e.g. _seconds)
	},
	latencyBuckets: prometheus.DefBuckets,
}
