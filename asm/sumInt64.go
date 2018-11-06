package asm

func SumInt64(x []int64) int64

func sumInt64(x []int64) (sum int64) {
	for _, v := range x {
		sum += v
	}
	return sum
}
