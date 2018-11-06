package asm

func Sxminusy(X, Y []float32)

func sxminusy(X, Y []float32) {
	for i := range X {
		Y[i] = X[i] - Y[i]
	}
}
