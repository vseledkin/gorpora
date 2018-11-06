package asm

func Sxminusyz(X, Y, Z []float32)

func sxminusyz(X, Y, Z []float32) {
	for i, x := range X {
		Z[i] = x - Y[i]
	}
}
