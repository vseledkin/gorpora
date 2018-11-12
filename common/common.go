package common

/*MaxInt max int*/
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

/*MinInt min int*/
func MinInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

/*MaxFloat32 max float32*/
func MaxFloat32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
