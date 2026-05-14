package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/Deimvis-go/xprometheus/prom/prommetric"
)

// TODO: move to promss
// TODO: rewrite RPSLatencyGroup as builder for IntervalMetricGroup
// - naming settings must be required
// - choose duration unit + suffix

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
			prommetric.NewHavingUnit(
				prometheus.V2.NewHistogramVec(prometheus.HistogramVecOpts{
					HistogramOpts: prometheus.HistogramOpts{
						Name:    baseName + cfg.naming.durationSuffix + cfg.naming.histogramSuffix,
						Buckets: cfg.latencyBuckets,
					},
					VariableLabels: finishLabels,
				}),
				prommetric.DurationToSecondsScaler,
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
	// namping: {base_name}{start/finish/duration suffix}{counter/hist suffix}
	startSuffix     string
	finishSuffix    string
	durationSuffix  string
	counterSuffix   string
	histogramSuffix string
}

var (
	// TODO: add base name to config and get rid of implicit defaults
	// (require explicit passing of default config)
	defaultRPSLatencyGroupConfig = rpsLatencyGroupCfg{
		naming: intervalMetricsNaming{
			startSuffix:     "_start",
			finishSuffix:    "_finish",
			durationSuffix:  "_duration",
			counterSuffix:   "_count",
			histogramSuffix: "", // NOTE: it's a good practice to leave a suffix with unit (e.g. _seconds)
		},
		latencyBuckets: prometheus.DefBuckets,
	}
)
