package utils

func StringSlicesEqual(l, r []string) bool {
	if len(l) != len(r) {
		return false
	}
	for i, v := range l {
		if r[i] != v {
			return false
		}
	}
	return true
}

func StringSlicesEqualIgnoreOrder(l, r []string) bool {
	if len(l) != len(r) {
		return false
	}

	lMap := map[string]bool{}
	for _, lv := range l {
		lMap[lv] = true
	}
	for _, rv := range r {
		if !lMap[rv] {
			return false
		}
	}
	return true
}
