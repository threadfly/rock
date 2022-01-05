package base

func MaxInt(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func MinInt(a, b int) (c int) {
	c = MaxInt(a, b)
	c = (a + b - c)
	return
}
