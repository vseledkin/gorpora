package asm

//Sset  set all components of a vector to a
func Sset(a float32, x []float32)

func sset(a float32, x []float32) {
	for i := range x {
		x[i] = a
	}
}
