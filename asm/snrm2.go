package asm

import "math"

//Snrm2 Euclidean norm: ||X||_2 = \sqrt {\sum X_i^2}
func Snrm2(X []float32) float32

func snrm2(X []float32) (nrm2 float32) {
	for _, x := range X {
		nrm2 += x * x
	}
	nrm2 = float32(math.Sqrt(float64(nrm2)))
	return
}
