package xprometheus

import (
	"fmt"
	"reflect"
)

func mustTrue(v bool, msg string) {
	if !v {
		panic(fmt.Errorf("xprometheus: assertion failed: %s", msg))
	}
}

func mustEq[T comparable](got T, want T, msg string) {
	if got != want {
		panic(fmt.Errorf("xprometheus: assertion failed: %s: got %v, want %v", msg, got, want))
	}
}

func mustImplement[Target any](v any) {
	targetT := reflect.TypeOf((*Target)(nil)).Elem()
	if v == nil {
		panic(fmt.Errorf("xprometheus: nil value does not implement %s", targetT.String()))
	}
	vt := reflect.TypeOf(v)
	if !vt.Implements(targetT) {
		panic(fmt.Errorf("xprometheus: type %s does not implement %s", vt.String(), targetT.String()))
	}
}
