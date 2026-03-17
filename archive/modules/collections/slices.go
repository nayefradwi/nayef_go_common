package collections

func MergeSlice[T any](slice1 []T, slice2 []T) []T {
	result := make([]T, 0)
	result = append(result, slice1...)
	result = append(result, slice2...)
	return result
}

func HasAllValues[T any, P string | int](values []T, allValues []T, getKey func(value T) P) bool {
	mapValues := make(map[P]bool)
	for _, value := range allValues {
		key := getKey(value)
		mapValues[key] = true
	}
	for _, value := range values {
		key := getKey(value)
		if !mapValues[key] {
			return false
		}
	}
	return true
}
