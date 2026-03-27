package errors

func Must[T any](f func() (T, error)) T {
	value, err := f()
	if err != nil {
		panic(err)
	}

	return value
}
