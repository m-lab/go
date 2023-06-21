package genericsx

// Contains returns whether the element `e` is in the slice `s`.
func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
