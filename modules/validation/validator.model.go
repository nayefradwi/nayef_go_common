package validation

import "github.com/nayefradwi/nayef_go_common/core"

type IValidator interface {
	Validate() error
}

type ValidateOption struct {
	Field   string
	Message string
	Data    any
}

func ValidateOptionsFrom(data any, field, message string) ValidateOption {
	return ValidateOption{Field: field, Message: message, Data: data}
}

type ValidationFunc func(opts ValidateOption) core.ErrorDetails

type Validator struct {
	Validations       []ValidationFunc
	ValidationOptions []ValidateOption
}

func NewValidator() *Validator {
	validator := &Validator{}
	validator.Validations = make([]ValidationFunc, 0)
	validator.ValidationOptions = make([]ValidateOption, 0)
	return validator
}

func (v *Validator) AddValidation(fn ValidationFunc, opts ValidateOption) {
	v.Validations = append(v.Validations, fn)
	v.ValidationOptions = append(v.ValidationOptions, opts)
}

func (v *Validator) Validate() error {
	errorDetails := make([]core.ErrorDetails, len(v.Validations))
	hasError := false

	for i, fn := range v.Validations {
		errorDetails[i] = fn(v.ValidationOptions[i])
		if errorDetails[i].Message != "" {
			hasError = true
		}
	}

	if hasError {
		return core.NewValidationError(errorDetails)
	}

	return nil
}
