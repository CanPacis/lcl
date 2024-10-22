package analyzer

func diff(slice1, slice2 []string) (excess, missing []string) {
	map1 := make(map[string]bool)
	map2 := make(map[string]bool)

	for _, v := range slice1 {
		map1[v] = true
	}

	for _, v := range slice2 {
		map2[v] = true
	}

	for _, v := range slice1 {
		if !map2[v] {
			excess = append(excess, v)
		}
	}

	for _, v := range slice2 {
		if !map1[v] {
			missing = append(missing, v)
		}
	}

	return
}
