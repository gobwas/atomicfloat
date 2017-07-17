package atomicfloat

import (
	"math"
	"sync/atomic"
	"unsafe"
)

type Float64 struct {
	value *float64
	store *uint64
}

func NewFloat64() Float64 {
	v := new(float64)
	s := (*uint64)(unsafe.Pointer(v))
	return Float64{v, s}
}

func (f Float64) Add(v float64) float64 {
	for i := 0; ; i++ {
		prev := *f.value
		next := prev + v
		if atomic.CompareAndSwapUint64(
			f.store,
			math.Float64bits(prev),
			math.Float64bits(next),
		) {
			return next
		}
	}
}

func (f Float64) Store(x float64) (prev float64) {
	bx := math.Float64bits(x)
	for {
		v := *f.value
		u := math.Float64bits(v)
		if atomic.CompareAndSwapUint64(f.store, u, bx) {
			return v
		}
	}
}

func (f Float64) Load() float64 {
	for {
		v := *f.value
		u := math.Float64bits(v)
		if atomic.CompareAndSwapUint64(f.store, u, u) {
			return v
		}
	}
}

func (f Float64) GreaterThanSwap(x float64) (swapped bool) {
	bx := math.Float64bits(x)
	for {
		v := *f.value
		u := math.Float64bits(v)

		var ux uint64
		if v > x {
			ux = bx
		} else {
			ux = u
		}
		if atomic.CompareAndSwapUint64(f.store, u, ux) {
			return ux == bx
		}
	}
}

func (f Float64) LessThanSwap(x float64) bool {
	bx := math.Float64bits(x)
	for {
		v := *f.value
		u := math.Float64bits(v)

		var ux uint64
		if v < x {
			ux = bx
		} else {
			ux = u
		}
		if atomic.CompareAndSwapUint64(f.store, u, ux) {
			return ux == bx
		}
	}
}
