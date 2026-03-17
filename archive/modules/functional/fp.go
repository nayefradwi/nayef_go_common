package functional

func Map[F any, T any](input []F, fn func(F) T) []T {
	result := make([]T, len(input))
	for i, value := range input {
		result[i] = fn(value)
	}
	return result
}

func Filter[T any](input []T, fn func(T) bool) []T {
	result := make([]T, 0)
	for _, value := range input {
		if fn(value) {
			result = append(result, value)
		}
	}
	return result
}

func FirstWhere[T any](input []T, fn func(T) bool) *T {
	for _, value := range input {
		if fn(value) {
			return &value
		}
	}
	return nil
}
