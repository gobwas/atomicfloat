package atomicfloat

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

func TestFloat64Add(t *testing.T) {
	for _, test := range []struct {
		add         []float64
		n           int
		parallelism int
	}{
		{
			add:         []float64{1, 2, -1},
			n:           1000000,
			parallelism: 1,
		},
		{
			add:         []float64{1, 2, -1},
			n:           1000000,
			parallelism: 8,
		},
		{
			add:         []float64{1, 2, -1},
			n:           100000,
			parallelism: 100,
		},
	} {
		name := fmt.Sprintf("%dx%dx%v", test.parallelism, test.n, test.add)

		t.Run(name, func(t *testing.T) {
			f := NewFloat64()

			var wg sync.WaitGroup
			for i := 0; i < test.parallelism; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < test.n; i++ {
						for _, v := range test.add {
							f.Add(v)
						}
					}
				}()
			}

			wg.Wait()

			var exp float64
			for i := 0; i < test.parallelism; i++ {
				for j := 0; j < test.n; j++ {
					for _, v := range test.add {
						exp += v
					}
				}
			}

			if act := f.Load(); act != exp {
				t.Fatalf("f.Load() = %v; want %v", act, exp)
			}
		})
	}
}

func TestFloat64MinMaxSwap(t *testing.T) {
	for _, test := range []struct {
		n           int
		parallelism int
	}{
		{
			n:           1000,
			parallelism: 1,
		},
		{
			n:           10000,
			parallelism: 8,
		},
		{
			n:           10000,
			parallelism: 1000,
		},
	} {
		name := fmt.Sprintf("%dx%d", test.parallelism, test.n)

		t.Run(name, func(t *testing.T) {
			fmin := NewFloat64()
			fmax := NewFloat64()

			total := test.n * test.parallelism
			min := float64(rand.Intn(1000000))
			max := min + float64(total-1)

			fmax.Store(min - 1)
			fmin.Store(max + 1)

			vs := make([]float64, total)
			for i := 0; i < total; i++ {
				vs[i] = min + float64(i)
			}
			perm := rand.Perm(total)

			var wg sync.WaitGroup
			for i := 0; i < test.parallelism; i++ {
				p := perm[test.n*i : test.n*(i+1)]

				wg.Add(1)
				go func() {
					defer wg.Done()
					for _, i := range p {
						fmin.GreaterThanSwap(vs[i])
						fmax.LessThanSwap(vs[i])
					}
				}()
			}

			wg.Wait()

			if actMax := fmax.Load(); actMax != max {
				t.Errorf("max f.Load() = %v; want %v", actMax, max)
			}
			if actMin := fmin.Load(); actMin != min {
				t.Errorf("min f.Load() = %v; want %v", actMin, min)
			}
		})
	}
}

func BenchmarkFloat64Add(b *testing.B) {
	b.Run("Float64", func(b *testing.B) {
		f := NewFloat64()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				f.Add(1)
			}
		})
	})
	b.Run("MutexAnalogue", func(b *testing.B) {
		c := &counter{}
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				c.Add(1)
			}
		})
	})
}

func BenchmarkFloati64GreaterThanSwap(b *testing.B) {
	b.Run("Float64", func(b *testing.B) {
		f := NewFloat64()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				f.GreaterThanSwap(1)
				f.Store(2)
			}
		})
	})
	b.Run("MutexAnalogue", func(b *testing.B) {
		c := &counter{}
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				c.GreaterThanSwap(1)
				c.Store(2)
			}
		})
	})
}

type counter struct {
	mu sync.Mutex
	v  float64
}

func (c *counter) Add(v float64) {
	c.mu.Lock()
	c.v += v
	c.mu.Unlock()
}

func (c *counter) Store(v float64) {
	c.mu.Lock()
	c.v = v
	c.mu.Unlock()
}

func (c *counter) GreaterThanSwap(v float64) {
	c.mu.Lock()
	if c.v > v {
		c.v = v
	}
	c.mu.Unlock()
}
