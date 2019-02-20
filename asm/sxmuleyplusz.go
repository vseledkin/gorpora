package asm

func Sxmuleyplusz(X, Y, Z []float32)

func sxmuleyplusz(X, Y, Z []float32) {
	for i := range X {
		Z[i] = Z[i] + X[i]*Y[i]
	}
}
