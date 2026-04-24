package xprommetric

import "time"

type UnitScaler[UnitT any, BaseUnitT any] interface {
	// Scale scales given value with unit
	// to value having base unit
	// (e.g. it may scale value having UnitT
	// described by time.Duration into value
	// of seconds described by float64).
	Scale(UnitT) (BaseUnitT, error)
	MustScale(UnitT) BaseUnitT
}

type SafeUnitScaler[UnitT any, BaseUnitT any] interface {
	SafeScale(UnitT) BaseUnitT
}

var DurationToSecondsScaler UnitScaler[time.Duration, float64] = durationToSecondsScaler{}

type durationToSecondsScaler struct{}

var _ UnitScaler[time.Duration, float64] = durationToSecondsScaler{}
var _ SafeUnitScaler[time.Duration, float64] = durationToSecondsScaler{}

func (s durationToSecondsScaler) Scale(d time.Duration) (float64, error) {
	return d.Seconds(), nil
}

func (s durationToSecondsScaler) MustScale(d time.Duration) float64 {
	return d.Seconds()
}

func (s durationToSecondsScaler) SafeScale(d time.Duration) float64 {
	return d.Seconds()
}
