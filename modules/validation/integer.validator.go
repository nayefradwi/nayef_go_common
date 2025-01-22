package validation

import (
	"strconv"

	"github.com/nayefradwi/nayef_go_common/core"
)

// at this point i got lazy and made chatgpt generate the rest of the code for me

type IntegerValidator struct {
	Validator *Validator
}

func NewIntegerValidator() *IntegerValidator {
	return &IntegerValidator{Validator: NewValidator()}
}

func IntegerValidatorFromValidator(validator *Validator) *IntegerValidator {
	return &IntegerValidator{Validator: validator}
}

func (i *IntegerValidator) MinValue(opts ValidateOption, min int) {
	vf := func(opts ValidateOption) core.ErrorDetails {
		num, errDetails := i.parseData(opts)
		if errDetails.Message != "" {
			return errDetails
		}

		if num < min {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	i.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}

func (i *IntegerValidator) MaxValue(opts ValidateOption, max int) {
	vf := func(opts ValidateOption) core.ErrorDetails {
		num, errDetails := i.parseData(opts)
		if errDetails.Message != "" {
			return errDetails
		}

		if num > max {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	i.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}

func (i *IntegerValidator) Between(opts ValidateOption, min, max int) {
	vf := func(opts ValidateOption) core.ErrorDetails {
		num, errDetails := i.parseData(opts)
		if errDetails.Message != "" {
			return errDetails
		}

		if num < min || num > max {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	i.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}

func (i *IntegerValidator) parseData(opts ValidateOption) (int, core.ErrorDetails) {
	switch v := opts.Data.(type) {
	case int:
		return v, core.ErrorDetails{}
	case *int:
		if v == nil {
			return 0, opts.ToInvalidDataType()
		}
		return *v, core.ErrorDetails{}
	case string:
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return 0, opts.ToInvalidDataType()
		}
		return parsed, core.ErrorDetails{}
	case *string:
		if v == nil {
			return 0, opts.ToInvalidDataType()
		}
		parsed, err := strconv.Atoi(*v)
		if err != nil {
			return 0, opts.ToInvalidDataType()
		}
		return parsed, core.ErrorDetails{}
	default:
		return 0, opts.ToInvalidDataType()
	}
}
