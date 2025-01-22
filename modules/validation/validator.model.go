package validation

import "github.com/nayefradwi/nayef_go_common/core"

const (
	INVALID_DATA_TYPE = "INVALID_DATA_TYPE"
)

type IValidator interface {
	Validate() error
}

type ValidateOption struct {
	Field   string
	Message string
	Data    any
}

func (opts ValidateOption) ToInvalidDataType() core.ErrorDetails {
	return core.ErrorDetails{Field: opts.Field, Message: "Invalid data type", Code: INVALID_DATA_TYPE}
}

func ValidateOptionsFrom(data any, field, message string) ValidateOption {
	return ValidateOption{Field: field, Message: message, Data: data}
}

type ValidationFunc struct {
	Opts ValidateOption
	fn   func(opts ValidateOption) core.ErrorDetails
}

func (vf ValidationFunc) Validate() core.ErrorDetails {
	return vf.fn(vf.Opts)
}

type Validator struct {
	Validations []ValidationFunc
}

func NewValidator() *Validator {
	validator := &Validator{}
	validator.Validations = make([]ValidationFunc, 0)
	return validator
}

func (v *Validator) AddValidation(fn ValidationFunc) {
	v.Validations = append(v.Validations, fn)
}

func (v *Validator) Validate() error {
	errorDetails := make([]core.ErrorDetails, len(v.Validations))
	hasError := false

	for i, fn := range v.Validations {
		errorDetails[i] = fn.Validate()
		if errorDetails[i].Message != "" {
			hasError = true
		}
	}

	if hasError {
		return core.NewValidationError(errorDetails)
	}

	return nil
}
