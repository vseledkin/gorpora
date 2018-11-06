package asm

// Snrm square of the norm
func Snrm(X []float32) float32

func snrm(X []float32) (nrm float32) {
	for _, x := range X {
		nrm += x * x
	}
	return
}
