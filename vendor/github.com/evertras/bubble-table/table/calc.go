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

// These var names are fine for this little function
//
//nolint:varnamelen
func gcd(x, y int) int {
	if x == 0 {
		return y
	} else if y == 0 {
		return x
	}

	return gcd(y%x, x)
}
