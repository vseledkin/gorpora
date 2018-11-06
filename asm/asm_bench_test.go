package asm

import (
	"math/rand"
	"testing"
)

func BenchmarkDxmulelyplusz(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float64, 1000000)
	y := make([]float64, 1000000)
	z := make([]float64, 1000000)
	for i := range x {
		x[i] = rand.Float64()*.2 - .1
		y[i] = rand.Float64()*.2 - .1
		z[i] = rand.Float64()*.2 - .1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		dxmulelyplusz(x, y, z)
	}
}

func BenchmarkDxmulelypluszOptimized(b *testing.B) {
	b.StopTimer()

	x := make([]float64, 1000000)
	y := make([]float64, 1000000)
	z := make([]float64, 1000000)
	for i := range x {
		x[i] = rand.Float64()*.2 - .1
		y[i] = rand.Float64()*.2 - .1
		z[i] = rand.Float64()*.2 - .1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		Dxmulelyplusz(x, y, z)
	}
}

func BenchmarkSxmulelyplusz(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	y := make([]float32, 1000000)
	z := make([]float32, 1000000)
	for i := range x {
		x[i] = rand.Float32()*.2 - .1
		y[i] = rand.Float32()*.2 - .1
		z[i] = rand.Float32()*.2 - .1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		sxmulelyplusz(x, y, z)
	}
}

func BenchmarkSxmulelypluszOptimized(b *testing.B) {
	b.StopTimer()

	x := make([]float32, 1000000)
	y := make([]float32, 1000000)
	z := make([]float32, 1000000)
	for i := range x {
		x[i] = rand.Float32()*.2 - .1
		y[i] = rand.Float32()*.2 - .1
		z[i] = rand.Float32()*.2 - .1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		Sxmulelyplusz(x, y, z)
	}
}

func BenchmarkSxminusyz(b *testing.B) {
	b.StopTimer()

	x := make([]float32, 1000000)
	y := make([]float32, 1000000)
	z := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
		y[i] = .2
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		sxminusyz(x, y, z)
	}
}

func BenchmarkSxminusyzOptimized(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	y := make([]float32, 1000000)
	z := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
		y[i] = .2
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		Sxminusyz(x, y, z)
	}
}

func BenchmarkSxminusy(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	y := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
		y[i] = .2
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		sxminusy(x, y)
	}
}

func BenchmarkSxminusyOptimized(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	y := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
		y[i] = .2
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		Sxminusy(x, y)
	}
}

func BenchmarkSxmulely(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	y := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
		y[i] = .2
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		sxmulely(x, y)
	}
}

func BenchmarkSxmulelyOptimized(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	y := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
		y[i] = .2
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		Sxmulely(x, y)
	}
}

func BenchmarkSset(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		sset(2.1, x)
	}
}

func BenchmarkSsetOptimized(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		Sset(2.1, x)
	}
}

func BenchmarkSxpy(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	y := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
		y[i] = .2
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		sxpy(x, y)
	}
}

func BenchmarkSxpyOptimized(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	y := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
		y[i] = .2
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		Sxpy(x, y)
	}
}

func BenchmarkSnrm(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		snrm(x)
	}
}

func BenchmarkOptimizedSnrm(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		Snrm(x)
	}
}

func BenchmarkSnrm2(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		snrm2(x)
	}
}

func BenchmarkOptimizedSnrm2(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		Snrm2(x)
	}
}

func BenchmarkSdot(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	y := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
		y[i] = .2
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		sdot(x, y)
	}
}

func BenchmarkOptimizedSdot(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	y := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
		y[i] = .2
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		Sdot(x, y)
	}
}

func BenchmarkSclean(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		sclean(x)
	}
}

func BenchmarkOptimizedSclean(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		Sclean(x)
	}
}

/*
func BenchmarkAsmInt642Sum(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	vector := make([]int64, 1000000)
	for i := range vector {
		vector[i] = 1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		SumInt642(vector)
	}
}*/
/*
func BenchmarkInt64Sum(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	vector := make([]int64, 1000000)
	for i := range vector {
		vector[i] = 1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		sumInt64(vector)
	}
}

func BenchmarkOptimizedInt64Sum(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	vector := make([]int64, 1000000)
	for i := range vector {
		vector[i] = 1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		SumInt64(vector)
	}
}

*/
func BenchmarkSsum(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	vector := make([]float32, 1000000)
	for i := range vector {
		vector[i] = 1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		ssum(vector)
	}
}

func BenchmarkOptimizedSsum(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	vector := make([]float32, 1000000)
	for i := range vector {
		vector[i] = 1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		Ssum(vector)
	}
}

func BenchmarkSaxpy(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	y := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
		y[i] = .2
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		saxpy(0.5, x, y)
	}
}

func BenchmarkOptimizedSaxpy(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	y := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
		y[i] = .2
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		Saxpy(0.5, x, y)
	}
}

func BenchmarkSscale(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		sscale(0.5, x)
	}
}

func BenchmarkOptimizedSscale(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()

	x := make([]float32, 1000000)
	for i := range x {
		x[i] = .1
	}
	b.StartTimer() //restart timer
	for i := 0; i < b.N; i++ {
		Sscale(0.5, x)
	}
}
