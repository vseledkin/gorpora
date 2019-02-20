package asm

func Sxmuley(X, Y []float32)

func sxmuley(X, Y []float32) {
	for i := range X {
		Y[i] *= X[i]
	}
}
