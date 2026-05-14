package prommetric

import "github.com/prometheus/client_golang/prometheus"

type HavingUnit[
	MetricT prometheus.Collector,
	UnitT any,
	BaseUnitT any,
] interface {
	prometheus.Collector
	// By convention, BaseUnitT should be accepted
	// by underlying metric's "observe" methods.
	UnitScaler[UnitT, BaseUnitT]

	Metric() MetricT
	Scaler() UnitScaler[UnitT, BaseUnitT]
}

func NewHavingUnit[
	MetricT prometheus.Collector,
	UnitT any,
	BaseUnitT any,
	UnitScalerT UnitScaler[UnitT, BaseUnitT],
](
	m MetricT,
	s UnitScalerT,
) HavingUnit[MetricT, UnitT, BaseUnitT] {
	return havingUnit[MetricT, UnitT, BaseUnitT, UnitScalerT]{
		m:          m,
		s:          s,
		Collector:  m,
		UnitScaler: s,
	}
}

type havingUnit[
	MetricT prometheus.Collector,
	UnitT any,
	BaseUnitT any,
	UnitScalerT UnitScaler[UnitT, BaseUnitT],
] struct {
	m MetricT
	s UnitScalerT
	prometheus.Collector
	UnitScaler[UnitT, BaseUnitT]
}

var _ HavingUnit[prometheus.Collector, any, any] = havingUnit[
	prometheus.Collector, any, any, UnitScaler[any, any],
]{}

func (hu havingUnit[MetricT, UnitT, BaseUnitT, UnitScalerT]) Metric() MetricT {
	return hu.m
}

func (hu havingUnit[MetricT, UnitT, BaseUnitT, UnitScalerT]) Scaler() UnitScaler[UnitT, BaseUnitT] {
	return hu.s
}
