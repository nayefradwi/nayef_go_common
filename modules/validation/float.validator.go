package validation

import (
	"strconv"

	"github.com/nayefradwi/nayef_go_common/core"
)

// at this point i got lazy and made chatgpt generate the rest of the code for me

type FloatValidator struct {
	Validator *Validator
}

func NewFloatValidator() *FloatValidator {
	return &FloatValidator{Validator: NewValidator()}
}

func FloatValidatorFromValidator(validator *Validator) *FloatValidator {
	return &FloatValidator{Validator: validator}
}

func (f *FloatValidator) MinValue(opts ValidateOption, min float64) {
	vf := func(opts ValidateOption) core.ErrorDetails {
		num, errDetails := f.parseData(opts)
		if errDetails.Message != "" {
			return errDetails
		}

		if num < min {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	f.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}

func (f *FloatValidator) MaxValue(opts ValidateOption, max float64) {
	vf := func(opts ValidateOption) core.ErrorDetails {
		num, errDetails := f.parseData(opts)
		if errDetails.Message != "" {
			return errDetails
		}

		if num > max {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	f.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}

func (f *FloatValidator) Between(opts ValidateOption, min, max float64) {
	vf := func(opts ValidateOption) core.ErrorDetails {
		num, errDetails := f.parseData(opts)
		if errDetails.Message != "" {
			return errDetails
		}

		if num < min || num > max {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	f.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}

func (f *FloatValidator) parseData(opts ValidateOption) (float64, core.ErrorDetails) {
	switch v := opts.Data.(type) {
	case float64:
		return v, core.ErrorDetails{}
	case *float64:
		if v == nil {
			return 0, opts.ToInvalidDataType()
		}
		return *v, core.ErrorDetails{}
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, opts.ToInvalidDataType()
		}
		return parsed, core.ErrorDetails{}
	case *string:
		if v == nil {
			return 0, opts.ToInvalidDataType()
		}
		parsed, err := strconv.ParseFloat(*v, 64)
		if err != nil {
			return 0, opts.ToInvalidDataType()
		}
		return parsed, core.ErrorDetails{}
	default:
		return 0, opts.ToInvalidDataType()
	}
}
