package validation

import "github.com/nayefradwi/nayef_go_common/core"

type SliceValidator struct {
	*Validator
}

func NewSliceValidator() *SliceValidator {
	return &SliceValidator{Validator: NewValidator()}
}

func SliceValidatorFromValidator(validator *Validator) *SliceValidator {
	return &SliceValidator{Validator: validator}
}

func (s *SliceValidator) NotEmpty(opts ValidateOption) {
	vf := func(opts ValidateOption) core.ErrorDetails {
		if opts.Data == nil {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		slice, ok := opts.Data.([]any)
		if !ok {
			return core.ErrorDetails{Field: opts.Field, Message: "Invalid data type", Code: INVALID_DATA_TYPE}
		}

		if len(slice) == 0 {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	s.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}
