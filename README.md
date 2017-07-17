# Atomic Float for Go

## Overview

This atomic float64 is slower than implementations with `sync.Mutex` for atomic
counters, but can be faster for compare and swap cases.

See benchmarks for more info:

```bash
go test -run=none -bench=. -cpu=4,8,12
```

## Usage

```go
package main

import "github.com/gobwas/atomicfloat"

func main() {
	f64 := atomicfloat.NewFloat64()
	
	for i := 0; i < 100; i++ {
		go f64.Add(1)
	}
	for i := 0; i < 58; i++ {
		go f64.Add(-1)
	}

	// Let all goroutines complete.
	<-time.After(time.Second)

	println(f64.Load()) // 42
}
```
