package util

func Filter[T any](arr []T, method func(item T) bool) []T {
	filtered := []T{}
	for _, a := range arr {
		if method(a) {
			filtered = append(filtered, a)
		}
	}
	return filtered
}
