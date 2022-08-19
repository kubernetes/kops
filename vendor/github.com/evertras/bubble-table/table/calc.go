package table

func min(x, y int) int {
	if x < y {
		return x
	}

	return y
}

func max(x, y int) int {
	if x > y {
		return x
	}

	return y
}

// nolint: varnamelen
func gcd(x, y int) int {
	if x == 0 {
		return y
	} else if y == 0 {
		return x
	}

	return gcd(y%x, x)
}
