package xprometheus

import (
	"errors"
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestScanStruct(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		type metrics struct{}
		smg := &StructMetricGroup[metrics]{}
		require.Len(t, smg.Collectors(), 0)
		smg.ScanStruct()
		require.Len(t, smg.Collectors(), 0)
		require.Len(t, smg.scanErrs, 0)
	})
	t.Run("one-metric", func(t *testing.T) {
		tcs := []struct {
			title           string
			oneMetricStruct interface{ Metric() prometheus.Collector }
		}{
			{
				"gauge",
				oneMetric[prometheus.Gauge]{
					V: prometheus.NewGauge(prometheus.GaugeOpts{}),
				},
			},
			{
				"gauge_vec",
				oneMetric[*prometheus.GaugeVec]{
					V: prometheus.NewGaugeVec(prometheus.GaugeOpts{}, []string{}),
				},
			},
			{
				"counter",
				oneMetric[prometheus.Counter]{
					V: prometheus.NewCounter(prometheus.CounterOpts{}),
				},
			},
			{
				"counter_vec",
				oneMetric[*prometheus.CounterVec]{
					V: prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{}),
				},
			},
			{
				"histogram",
				oneMetric[prometheus.Histogram]{
					V: prometheus.NewHistogram(prometheus.HistogramOpts{}),
				},
			},
			{
				"histogram_vec",
				oneMetric[*prometheus.HistogramVec]{
					V: prometheus.NewHistogramVec(prometheus.HistogramOpts{}, []string{}),
				},
			},
			{
				"summary",
				oneMetric[prometheus.Summary]{
					V: prometheus.NewSummary(prometheus.SummaryOpts{}),
				},
			},
			{
				"summary_vec",
				oneMetric[*prometheus.SummaryVec]{
					V: prometheus.NewSummaryVec(prometheus.SummaryOpts{}, []string{}),
				},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.title, func(t *testing.T) {
				smg := NewStructMetricGroup(tc.oneMetricStruct)
				smg.ScanStruct()
				require.Len(t, smg.Collectors(), 1)
				require.Equal(t, tc.oneMetricStruct.Metric(), smg.Collectors()[0])
			})
		}
	})
	t.Run("uninitialized-metrics-ignored", func(t *testing.T) {
		tcs := []struct {
			title           string
			oneMetricStruct interface{ Metric() prometheus.Collector }
		}{
			{
				"gauge",
				oneMetric[prometheus.Gauge]{},
			},
			{
				"gauge_vec",
				oneMetric[*prometheus.GaugeVec]{},
			},
			{
				"counter",
				oneMetric[prometheus.Counter]{},
			},
			{
				"counter_vec",
				oneMetric[*prometheus.CounterVec]{},
			},
			{
				"histogram",
				oneMetric[prometheus.Histogram]{},
			},
			{
				"histogram_vec",
				oneMetric[*prometheus.HistogramVec]{},
			},
			{
				"summary",
				oneMetric[prometheus.Summary]{},
			},
			{
				"summary_vec",
				oneMetric[*prometheus.SummaryVec]{},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.title, func(t *testing.T) {
				smg := NewStructMetricGroup(tc.oneMetricStruct)
				smg.ScanStruct()
				require.Len(t, smg.Collectors(), 0)
				require.Len(t, smg.scanErrs, 0)
				err := smg.ValidateAllCollectorsInitialized()
				require.Error(t, err)
			})
		}
	})
	t.Run("ignore-unrecognized-fields", func(t *testing.T) {
		type metrics struct {
			Metric    prometheus.Counter
			NonMetric int64
		}
		smg := NewStructMetricGroup(metrics{
			Metric:    prometheus.NewCounter(prometheus.CounterOpts{}),
			NonMetric: 42,
		})
		smg.ScanStruct()
		require.Len(t, smg.Collectors(), 1)
		require.Len(t, smg.scanErrs, 1)
		require.ErrorAs(t, smg.scanErrs[0], &unrecognizedFieldError{})
		err := smg.ValidateAllFieldsHandled()
		require.Error(t, err)
	})
	t.Run("post-metrics-initialization", func(t *testing.T) {
		smg := NewStructMetricGroup(oneMetric[prometheus.Counter]{
			V: nil,
		})
		smg.ScanStruct()
		require.Len(t, smg.Collectors(), 0)
		err := smg.ValidateAllCollectorsInitialized()
		require.Error(t, err)

		smg.Struct().V = prometheus.NewCounter(prometheus.CounterOpts{})
		require.Len(t, smg.Collectors(), 1)
		err = smg.ValidateAllCollectorsInitialized()
		require.NoError(t, err)
	})
	t.Run("subgroups", func(t *testing.T) {
		t.Run("collectors-propagated", func(t *testing.T) {
			type metrics struct {
				Sub oneMetric[prometheus.Counter]
			}
			smg := NewStructMetricGroup(metrics{
				Sub: oneMetric[prometheus.Counter]{
					V: prometheus.NewCounter(prometheus.CounterOpts{}),
				},
			})
			smg.ScanStruct()
			require.Len(t, smg.Collectors(), 1)
			err := smg.ValidateAllFieldsHandled()
			require.NoError(t, err)
			err = smg.ValidateAllCollectorsInitialized()
			require.NoError(t, err)
		})
		t.Run("init-errors-propagated", func(t *testing.T) {
			type metrics struct {
				Sub oneMetric[prometheus.Counter]
			}
			smg := NewStructMetricGroup(metrics{
				Sub: oneMetric[prometheus.Counter]{
					V: nil,
				},
			})
			smg.ScanStruct()
			require.Len(t, smg.Collectors(), 0)
			err := smg.ValidateAllFieldsHandled()
			require.NoError(t, err)
			err = smg.ValidateAllCollectorsInitialized()
			require.Error(t, err)
		})
		t.Run("uninitialized-ignored", func(t *testing.T) {
			t.Run("direct-impl", func(t *testing.T) {
				type metrics struct {
					Sub MetricGroup
				}
				smg := NewStructMetricGroup(metrics{
					Sub: (*oneMetric[prometheus.Counter])(nil),
				})
				smg.ScanStruct()
				require.Len(t, smg.Collectors(), 0)
				err := smg.ValidateAllFieldsHandled()
				require.NoError(t, err)
				err = smg.ValidateAllCollectorsInitialized()
				require.Error(t, err)
			})
			t.Run("impl-by-embedding", func(t *testing.T) {
				type metricGroup struct {
					*oneMetric[prometheus.Counter]
				}
				type metrics struct {
					Sub metricGroup
				}
				smg := NewStructMetricGroup(metrics{
					Sub: metricGroup{},
					// metric group is not empty,
					// but embedded struct, which
					// implements MetricGroup,
					// is nil
				})
				smg.ScanStruct()
				require.Len(t, smg.Collectors(), 0)
				err := smg.ValidateAllFieldsHandled()
				require.NoError(t, err)
				err = smg.ValidateAllCollectorsInitialized()
				require.Error(t, err)
			})
		})
	})
}

type oneMetric[T prometheus.Collector] struct {
	V T
}

var _ MetricGroup = oneMetric[prometheus.Counter]{}
var _ MetricGroup = (*oneMetric[prometheus.Counter])(nil)

func (m oneMetric[T]) Metric() prometheus.Collector {
	return m.V
}

func (m oneMetric[T]) Collectors() []prometheus.Collector {
	var crs []prometheus.Collector
	rv := reflect.ValueOf(m.V)
	if rv.IsValid() && !rv.IsZero() && !rv.IsNil() {
		crs = append(crs, any(m.V).(prometheus.Collector))
	}
	return crs
}

func (m oneMetric[T]) ValidateAllCollectorsInitialized() error {
	rv := reflect.ValueOf(m.V)
	if rv.IsValid() && !rv.IsZero() && !rv.IsNil() {
		return nil
	}
	return errors.New("metric not initialized")
}
