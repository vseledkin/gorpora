package asm

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestMatrixVectorProduct(t *testing.T) {
	want := []float32{8, 18, 28}
	m := NewMatrix32FromArray([][]float32{
		{1, 2},
		{3, 4},
		{5, 6}})

	v := []float32{2, 3}
	got := []float32{2, 3, 4}
	m.Mult(v, got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Matrix vector product failed want %v got %v\n", want, got)
	}
}

func TestMatrixShift(t *testing.T) {
	want := NewMatrix32FromArray([][]float32{
		{5, 8},
		{9, 13},
		{13, 18}})
	//wanttransposed = want.T()

	m := NewMatrix32FromArray([][]float32{
		{1, 2},
		{3, 4},
		{5, 6}})

	right := []float32{2, 3}
	left := []float32{2, 3, 4}
	m.Shift(1.0, left, right)
	mt := m.T()
	m.t = nil
	// check matrix
	fmt.Println(m)
	fmt.Println(want)
	if !reflect.DeepEqual(m, want) {
		t.Fatalf("Matrix shift failed, want %v got %v\n", want, m)
	}
	// check transposed
	wantt := want.T()
	wantt.t = nil
	mt.t = nil
	fmt.Println(mt)
	fmt.Println(wantt)
	if !reflect.DeepEqual(wantt, mt) {
		t.Fatalf("Matrix shift failed, want %v got %v\n", wantt, mt)
	}
}

func TestVisitRowByRow(t *testing.T) {
	var columns uint32 = 2
	var rows uint32 = 3
	m := NewMatrix32(rows, columns)
	var count uint32
	m.VisitRowByRowElements(func(elp *float32) {
		count++
	})
	if count != rows*columns {
		t.Fatalf("failed to visit all elements want %d got %d in matrix (%d,%d)\n", rows*columns, count, m.rows, m.columns)
	}
}

func TestTranspose(t *testing.T) {
	m := NewRandomMatrix32(3, 2, -0.1, 0.1)
	mt := m.T()
	var i, j uint32
	for i = 0; i < m.rows; i++ {
		for j = 0; j < m.columns; j++ {
			if m.At(i, j) != mt.At(j, i) {
				t.Fatalf("Matrix transpose failed want %f got %f at (i:%d,j:%d)\n", m.At(i, j),
					mt.At(j, i), i, j)
			}
		}
	}
	fmt.Println(m)
	fmt.Println(mt)
}

func TestSxmulelyplusz(t *testing.T) {
	for j := 0; j < 100; j++ {
		X := make([]float32, j)
		Y := make([]float32, j)
		Z1 := make([]float32, j)
		Z2 := make([]float32, j)
		for i := range X {
			X[i] = rand.Float32()*.2 - .1
			Y[i] = rand.Float32()*.2 - .1
			Z1[i] = rand.Float32()*.2 - .1
			Z2[i] = Z1[i]
		}
		Sxmulelyplusz(X, Y, Z1)
		sxmulelyplusz(X, Y, Z2)
		for i := range Y {
			if Z1[i] != Z2[i] {
				t.Fatalf("product do not match want %e got %e in vector of length %d\n", Z1[i], Z2[i], len(Y))
			}
		}
	}
}

func TestDxmulelyplusz(t *testing.T) {
	for j := 0; j < 100; j++ {
		X := make([]float64, j)
		Y := make([]float64, j)
		Z1 := make([]float64, j)
		Z2 := make([]float64, j)
		for i := range X {
			X[i] = rand.Float64()*.2 - .1
			Y[i] = rand.Float64()*.2 - .1
			Z1[i] = rand.Float64()*.2 - .1
			Z2[i] = Z1[i]
		}
		Dxmulelyplusz(X, Y, Z1)
		dxmulelyplusz(X, Y, Z2)
		for i := range Y {
			if Z1[i] != Z2[i] {
				t.Fatalf("product do not match want %.20f got %.20f in vector of length %d\n", Z1[i], Z2[i], len(Y))
			}
		}
	}
}

func TestSxminusyz(t *testing.T) {
	for j := 0; j < 100; j++ {
		X := make([]float32, j)
		Y := make([]float32, j)
		Z1 := make([]float32, j)
		Z2 := make([]float32, j)
		for i := range X {
			X[i] = rand.Float32()*.2 - .1
			Y[i] = rand.Float32()*.2 - .1
		}
		sxminusyz(X, Y, Z1)
		Sxminusyz(X, Y, Z2)
		for i := range Y {
			if Z1[i] != Z2[i] {
				t.Fatalf("product do not match want %f got %f in vector of length %d\n", Z1[i], Z2[i], len(Y))
			}
		}
	}
}

func TestSxmulely(t *testing.T) {
	for j := 0; j < 100; j++ {
		X := make([]float32, j)
		Y := make([]float32, j)
		Y1 := make([]float32, j)
		for i := range X {
			X[i] = rand.Float32()*.2 - .1
			Y[i] = rand.Float32()*.2 - .1
			Y1[i] = Y[i]
		}
		sxmulely(X, Y)
		Sxmulely(X, Y1)
		for i := range Y {
			if Y[i] != Y1[i] {
				t.Fatalf("product do not match want %f got %f in vector of length %d\n", Y[i], Y1[i], len(Y))
			}
		}
	}
}

func TestSxminusy(t *testing.T) {
	for j := 0; j < 100; j++ {
		X := make([]float32, j)
		Y := make([]float32, j)
		Y1 := make([]float32, j)
		for i := range X {
			X[i] = rand.Float32()*.2 - .1
			Y[i] = rand.Float32()*.2 - .1
			Y1[i] = Y[i]
		}
		sxminusy(X, Y)
		Sxminusy(X, Y1)
		for i := range Y {
			if Y[i] != Y1[i] {
				t.Fatalf("sums do not match want %f got %f in vector of length %d\n", Y[i], Y1[i], len(Y))
			}
		}
	}
}

func TestSset(t *testing.T) {
	for j := 0; j < 100; j++ {
		X := make([]float32, j)
		X1 := make([]float32, j)
		for i := range X {
			X[i] = .1
			X1[i] = .1
		}
		sset(2.1, X)
		Sset(2.1, X1)
		for i := range X {
			if X[i] != X1[i] {
				t.Fatalf("values do not match want %f got %f in vector of length %d\n", X[i], X1[i], len(X))
			}
			if X[i] != 2.1 {
				t.Fatalf("values do not match want %f got %f in vector of length %d\n", 2.1, X1[i], len(X))
			}
		}
	}
}

func TestSxpy(t *testing.T) {
	for j := 0; j < 100; j++ {
		X := make([]float32, j)
		Y := make([]float32, j)
		Y1 := make([]float32, j)
		for i := range X {
			X[i] = .1
			Y[i] = .1
			Y1[i] = .1
		}
		sxpy(X, Y)
		Sxpy(X, Y1)
		for i := range Y {
			if Y[i] != Y1[i] {
				t.Fatalf("sums do not match want %f got %f in vector of length %d\n", Y[i], Y1[i], len(Y))
			}
		}
	}
}

func TestSclean(t *testing.T) {
	for j := 0; j < 100; j++ {
		X := make([]float32, j)
		X1 := make([]float32, j)
		for i := range X {
			X[i] = rand.Float32()*0.2 - 1
			X1[i] = rand.Float32()*0.2 - 1
		}
		sclean(X)
		//fmt.Println(X1)
		Sclean(X1)
		//fmt.Println(X1)
		for i := range X {
			if X[i] != X1[i] || X1[i] != 0 || X[i] != 0 {
				t.Fatalf("cleaned do not match want %f got %f in vector of length %d\n", X[i], X1[i], len(X))
			}
		}
	}
}

func TestSnrm2(t *testing.T) {
	for j := 0; j < 100; j++ {
		X := make([]float32, j)
		for i := range X {
			X[i] = .1
		}
		dot := snrm2(X)
		dot1 := Snrm2(X)
		if dot-dot1 > 0.000001 || dot-dot1 < -0.000001 {
			t.Fatalf("norm2 do not match want %f got %f in vector of length %d\n", dot, dot1, len(X))
		}
	}
}

func TestSnrm(t *testing.T) {
	for j := 0; j < 100; j++ {
		X := make([]float32, j)
		for i := range X {
			X[i] = .1
		}
		dot := snrm(X)
		dot1 := Snrm(X)
		if dot-dot1 > 0.000001 || dot-dot1 < -0.000001 {
			t.Fatalf("norm do not match want %f got %f in vector of length %d\n", dot, dot1, len(X))
		}
	}
}

func TestSdot(t *testing.T) {
	x := []float32{0.40724772, 0.0712502, 0.041903675, 0.15231317, 0.21472728, 0.4622725, -0.0903995, 0.24077353, 0.006599188, -0.47139943, 0.3086093, 0.1786874, 0.42446965, 0.22735131, 0.46515256}
	//y := []float32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	y := make([]float32, len(x))
	dot := sdot(x, y)
	dot1 := Sdot(x, y)

	if dot1 != 0 || dot != 0 || dot-dot1 > 0.00004 || dot-dot1 < -0.00004 {
		t.Fatalf("dot do not match want %f got %f in vector of length %d\n", dot, dot1, len(x))
	}

	for j := 0; j < 100; j++ {
		X := make([]float32, j)
		Y := make([]float32, j)
		for i := range X {
			X[i] = rand.Float32()*2. - 1.
			Y[i] = rand.Float32()*2. - 1.
		}

		for i := 0; i < 10; i++ {
			dot := sdot(X, Y)
			dot1 := Sdot(X, Y)
			if dot-dot1 > 0.00004 || dot-dot1 < -0.00004 {
				t.Fatalf("dot do not match want %f got %f in vector of length %d in test %d\n", dot, dot1, len(X), i)
			}
		}
	}
}

func TestInt64Sum(t *testing.T) {
	for j := 0; j < 100; j++ {
		vector := make([]int64, j)
		for i := range vector {
			vector[i] = 1
		}
		sumfast := SumInt64(vector)
		sum := sumInt64(vector)
		if sumfast != sum {
			t.Fatalf("sums do not match want %d got %d in vector of length %d\n", sum, sumfast, len(vector))
		}
		if int(sumfast) != j {
			t.Fatalf("sse wrong sum value %d expected %d in vector of length %d\n", sumfast, j, len(vector))
		}
		if int(sum) != j {
			t.Fatalf("wrong sum value %d expected %d in vector of length %d\n", sum, j, len(vector))
		}
	}
}

func TestSsum(t *testing.T) {
	for j := 0; j < 100; j++ {
		vector := make([]float32, j)
		for i := range vector {
			vector[i] = 1
		}
		sumfast := Ssum(vector)
		sum := ssum(vector)
		if sumfast != sum {
			t.Fatalf("sums do not match want %f got %f in vector of length %d\n", sum, sumfast, len(vector))
		}
		if int(sumfast) != j {
			t.Fatalf("sse wrong sum value %f expected %f in vector of length %d\n", sumfast, float32(j), len(vector))
		}
		if int(sum) != j {
			t.Fatalf("wrong sum value %f expected %f in vector of length %d\n", sum, float32(j), len(vector))
		}
	}
}

func TestSaxpy(t *testing.T) {
	for j := 0; j < 100; j++ {
		X := make([]float32, j)
		Y := make([]float32, j)
		Y1 := make([]float32, j)
		for i := range X {
			X[i] = rand.Float32()*0.2 - 0.1
			Y[i] = rand.Float32()*0.2 - 0.1
		}
		for i, y := range Y {
			Y1[i] = y
		}
		r := rand.Float32()*0.2 - 0.1
		saxpy(r, X, Y)
		//fmt.Println(Y1)
		Saxpy(r, X, Y1)
		//fmt.Println(Y1)
		for i := range Y {
			if Y[i] != Y1[i] {
				t.Fatalf("sums do not match want %f got %f in vector of length %d\n", Y[i], Y1[i], len(Y))
			}
		}
	}
}

func TestSscale(t *testing.T) {
	for j := 0; j < 100; j++ {
		X := make([]float32, j)
		X1 := make([]float32, j)
		Y := make([]float32, j)
		for i := range X {
			X[i] = .1
			X1[i] = .1
		}
		sscale(.5, X)
		Sscale(.5, X1)
		for i := range Y {
			if X[i] != X1[i] {
				t.Fatalf("sums do not match want %f got %f in vector of length %d\n", X[i], X1[i], len(X))
			}
		}
	}
}
