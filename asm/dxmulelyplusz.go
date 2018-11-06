package asm

func Dxmulelyplusz(X, Y, Z []float64)

func dxmulelyplusz(X, Y, Z []float64) {
	for i := range X {
		Z[i] = Z[i] + X[i]*Y[i]
	}
}
