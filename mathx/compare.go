package mathx

import "errors"

// Min returns the minimum of two ints.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Mode returns the mode of a slice.
func Mode(slice []int64) (int64, error) {
	if len(slice) == 0 {
		return 0, errors.New("invalid slice")
	}

	counts := make(map[int64]int64)
	var mode int64
	var maxCount int64

	for _, v := range slice {
		counts[v]++
		if counts[v] > maxCount {
			maxCount = counts[v]
			mode = v
		}
	}

	return mode, nil
}
