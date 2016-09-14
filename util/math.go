package util

func MinInt(a, b int) int {
	if a > b {
		return b
	} else {
		return a
	}
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
