package xprometheus

import (
	"net/http"
	"reflect"
)

// RoundTripFn is the functional form of [http.RoundTripper].
type RoundTripFn func(*http.Request) (*http.Response, error)

// RoundTripWrapFn wraps a [RoundTripFn] with extra behavior (e.g. metric recording).
type RoundTripWrapFn func(RoundTripFn) RoundTripFn

// PathNormalizer returns a normalized, low-cardinality path for a request.
// For example, "/users/sDaf34Fb9" may normalize to "/users/:id".
// It returns nil if no normalized path is available.
type PathNormalizer func(*http.Request) *string

func recursiveIndirect(v reflect.Value) reflect.Value {
	for {
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface:
			if v.IsNil() {
				return v
			}
			v = v.Elem()
		default:
			return v
		}
	}
}

func flatten[T any](s [][]T) []T {
	sz := 0
	for i := range s {
		sz += len(s[i])
	}
	res := make([]T, 0, sz)
	for i := range s {
		res = append(res, s[i]...)
	}
	return res
}

func mapSlice[T, U any](s []T, fn func(T) U) []U {
	res := make([]U, len(s))
	for i := range s {
		res[i] = fn(s[i])
	}
	return res
}

func filterSlice[T any](s []T, pred func(T) bool) []T {
	res := make([]T, 0, len(s))
	for i := range s {
		if pred(s[i]) {
			res = append(res, s[i])
		}
	}
	return res
}

func ptrDerefOr[T any](v *T, fallback T) T {
	if v != nil {
		return *v
	}
	return fallback
}
