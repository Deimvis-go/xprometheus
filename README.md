# xprometheus

Helpers for building Prometheus metrics: typed metric groups, struct-based metric registration, and an HTTP round-trip wrapper.

## Features

* `MetricGroup` interface and composable groups (`IntervalMetricGroup`, `CounterHistMetricGroup`, `RPSLatencyGroup`)
* `StructMetricGroup[T]` — declare metrics as struct fields and discover/register them via reflection
* `RoundTripWrapFn` that records RPS + latency for outgoing HTTP calls with optional path normalization
* Unit-aware metrics via `prommetric.HavingUnit` (e.g. histograms that observe `time.Duration` but export seconds)

## Quick Start

```golang
package quick_start

import (
    "time"

    "github.com/prometheus/client_golang/prometheus"

    "github.com/Deimvis-go/xprometheus/prom"
)

type MyMetrics struct {
    Requests *prometheus.CounterVec
}

func (m MyMetrics) Collectors() []prometheus.Collector {
    return []prometheus.Collector{m.Requests}
}

func main() {
    g := prom.NewRPSLatencyGroup(
        "my_service",
        prometheus.ConstrainableLabels{},
        prometheus.ConstrainableLabels{},
    )
    g.Record(prometheus.Labels{}, func(rc prom.RecordControl) {
        time.Sleep(10 * time.Millisecond)
    })
    prometheus.MustRegister(g.Collectors()...)
}
```
