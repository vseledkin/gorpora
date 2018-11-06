package asm

func Sxmulelyplusz(X, Y, Z []float32)

func sxmulelyplusz(X, Y, Z []float32) {
	for i := range X {
		Z[i] = Z[i] + X[i]*Y[i]
	}
}
