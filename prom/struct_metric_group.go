package prom

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/Deimvis/go-ext/go1.25/ext"
	"github.com/Deimvis/go-ext/go1.25/xcheck/xmust"
	"github.com/Deimvis/go-ext/go1.25/xreflect"
	"github.com/Deimvis/go-ext/go1.25/xslices"
)

// TODO: support auto creation of metrics:
// specify autofill and specify labels in tags
// - tag example: `metric:"autoinit(ls=['label1', 'label2'])"`
// - New and ScanStruct should accept option IgnoreMetricsAutoinit

func NewStructMetricGroup[T any](v T) *StructMetricGroup[T] {
	return &StructMetricGroup[T]{v: v}
}

type StructMetricGroup[T any] struct {
	v T

	// getters are used because
	// collectors are struct fields
	// and may be set in anytime.
	selfCrsGetters   []func() (prometheus.Collector, error)
	subgroupsGetters []func() (MetricGroup, error)
	scanOnce         sync.Once
	scanErrs         []error
}

var _ MetricGroup = (*StructMetricGroup[any])(nil)

func (smg *StructMetricGroup[T]) Struct() *T {
	return &smg.v
}

// Clone creates a copy,
// which operates over the same metrics,
// but may store them in different state
// (e.g. with some labels precompiled).
// Clone is primarily useful for independent
// recording of different events,
// whose stats should be exported
// to the same set of metrics.
func (smg *StructMetricGroup[T]) Clone() *StructMetricGroup[T] {
	return &StructMetricGroup[T]{
		v: smg.v,
		// can't copy scan results:
		// - scan can be in progress right now,
		//   so we need to invoke scan here,
		//   which may be not a good idea
		// - scan results have pointers to struct fields,
		//   but we copy the struct and pointers would be invalid
	}
}

// ScanStruct allows to explicitly call struct scanning.
// It is unnecessary, because scan will implicitly
// when it's needed.
func (smg *StructMetricGroup[T]) ScanStruct() []error {
	smg.scanCollectorsOnce()
	return smg.scanErrs
}

// Collectors returns collectors that were found
// after scanning internal struct and
// only those that are not nil.
func (smg *StructMetricGroup[T]) Collectors() []prometheus.Collector {
	smg.scanCollectorsOnce()
	return slices.Concat(
		smg.SelfCollectors(),
		xslices.Flatten(ext.Map(smg.subgroups(), MetricGroup.Collectors)),
	)
}

// SelfCollectors returns collectors that were found
// after scanning internal struct and
// only those that are not nil and located directly
// in the struct fields
// (without recursion to fields that are groups).
func (smg *StructMetricGroup[T]) SelfCollectors() []prometheus.Collector {
	smg.scanCollectorsOnce()
	var crs []prometheus.Collector
	for _, get := range smg.selfCrsGetters {
		cr, err := get()
		if err == nil {
			crs = append(crs, cr)
		}
	}
	return crs
}

func (smg *StructMetricGroup[T]) subgroups() []MetricGroup {
	smg.scanCollectorsOnce()
	var sgs []MetricGroup
	for _, get := range smg.subgroupsGetters {
		sg, err := get()
		if err == nil {
			sgs = append(sgs, sg)
		}
	}
	return sgs
}

// ValidateAllFieldsHandled checks whether all fields of T
// are recognized and handled, otherwise non-nil error is returned.
// It works recursively for fields that implement the same method.
func (smg *StructMetricGroup[T]) ValidateAllFieldsHandled() error {
	smg.scanCollectorsOnce()
	ufErrs := ext.Filter(smg.scanErrs, func(err error) bool {
		ufErr := unrecognizedFieldError{}
		return errors.As(err, &ufErr)
	})
	if len(ufErrs) > 0 {
		return errors.Join(ufErrs...)
	}
	return nil
}

// ValidateAllCollectorsHaveValue checks whether all known
// collectors are initialized, i.e., have value.
// It works recursively for fields that implement the same method.
func (smg *StructMetricGroup[T]) ValidateAllCollectorsInitialized() error {
	smg.scanCollectorsOnce()
	var errs []error
	for _, get := range smg.selfCrsGetters {
		_, err := get()
		if err != nil {
			errs = append(errs, err)
		}
	}
	for _, get := range smg.subgroupsGetters {
		sg, err := get()
		if err != nil {
			errs = append(errs, err)
		} else if sgv, ok := sg.(interface{ ValidateAllCollectorsInitialized() error }); ok {
			err := sgv.ValidateAllCollectorsInitialized()
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func (smg *StructMetricGroup[T]) scanCollectorsOnce() {
	smg.scanOnce.Do(smg.scanCollectors)
}

// scanCollectors scans struct fields
// and handles if field implements either
// prometheus.Collector or MetricGroup.
// If field stores nil value somewhere
// in the indirection chain, it will be
// assumed as uninitialized, even though it may
// work correctly (e.g. nil pointer to struct
// that implements prometheus.Collector and
// does no indirection, making it work properly
// - this will also be assumed as uninitialized
// and won't be included into Collectors() output).
func (smg *StructMetricGroup[T]) scanCollectors() {
	v := reflect.ValueOf(&smg.v).Elem()
	v = xreflect.RecursiveIndirect(v)
	xmust.Eq(v.Kind(), reflect.Struct,
		"StructMetricGroup type paramter's internal Kind must be Struct")
	vt := v.Type()
	for i := 0; i < vt.NumField(); i++ {
		sf := vt.Field(i)
		if sf.Anonymous {
			continue
		}
		if sf.Type.Implements(metricGroupT) {
			smg.subgroupsGetters = append(smg.subgroupsGetters,
				func(f reflect.Value, sf reflect.StructField) func() (MetricGroup, error) {
					return func() (MetricGroup, error) {
						err := checkImplementationNotNil(f, metricGroupT)
						if err != nil {
							return nil, fmt.Errorf("metric group field '%s': %w", sf.Name, err)
						}
						sg := f.Interface()
						if sg == nil {
							return nil, fmt.Errorf("metric group field '%s' is nil (%s)",
								sf.Name, f.Type().String())
						}
						return sg.(MetricGroup), nil
					}
				}(v.Field(i), sf),
			)
		} else if sf.Type.Implements(promCollectorT) {
			smg.selfCrsGetters = append(smg.selfCrsGetters,
				func(f reflect.Value, sf reflect.StructField) func() (prometheus.Collector, error) {
					return func() (prometheus.Collector, error) {
						err := checkImplementationNotNil(f, promCollectorT)
						if err != nil {
							return nil, fmt.Errorf("collector field '%s': %w", sf.Name, err)
						}
						cr := f.Interface()
						if cr == nil {
							return nil, fmt.Errorf("collector field '%s' is nil (%s)",
								sf.Name, f.Type().String())
						}
						return cr.(prometheus.Collector), nil
					}
				}(v.Field(i), sf),
			)
		} else {
			err := unrecognizedFieldError{sf: sf}
			smg.scanErrs = append(smg.scanErrs, err)
		}
	}
}

// checkImplementationNotNil recursively traverses struct
// and checks that final implementation of interfaceT
// has not nil value
// (struct may implement interface by embedding another value).
func checkImplementationNotNil(v reflect.Value, interfaceT reflect.Type) error {
	internalV := xreflect.RecursiveIndirect(v)
	for (internalV.Kind() == reflect.Pointer ||
		internalV.Kind() == reflect.Interface) &&
		internalV.IsNil() {
		return fmt.Errorf("nil implementation '%s' of interface '%s'",
			internalV.Type().String(), // because anonymous fields have Type().Name() empty
			interfaceT.Name())
	}
	if internalV.Kind() == reflect.Struct {
		for i := 0; i < internalV.NumField(); i++ {
			f := internalV.Field(i)
			sf := internalV.Type().Field(i)
			if sf.Anonymous {
				if sf.Type.Implements(interfaceT) {
					err := checkImplementationNotNil(f, interfaceT)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

type unrecognizedFieldError struct {
	sf reflect.StructField
}

var _ error = unrecognizedFieldError{}

func (ufr unrecognizedFieldError) Error() string {
	return fmt.Sprintf("unrecognized field '%s' (%s)", ufr.sf.Name, ufr.sf.Type.String())
}

var (
	promCollectorT = reflect.TypeOf((*prometheus.Collector)(nil)).Elem()
	metricGroupT   = reflect.TypeOf((*MetricGroup)(nil)).Elem()
)
