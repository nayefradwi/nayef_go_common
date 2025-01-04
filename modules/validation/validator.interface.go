package validation

type IValidator interface {
	Validate() error
}
type ValidateOptions struct {
	Field   string
	Message string
}

type ValidationFunc[T any] func(data T, errMsg string) error
