package validation

import (
	"time"

	"github.com/nayefradwi/nayef_go_common/core"
)

type DateValidator struct {
	*Validator
}

func NewDateValidator() *DateValidator {
	return &DateValidator{Validator: NewValidator()}
}

func DateValidatorFromValidator(validator *Validator) *DateValidator {
	return &DateValidator{Validator: validator}
}

func (d *DateValidator) IsDate(opts ValidateOption) {
	vf := func(opts ValidateOption) core.ErrorDetails {
		_, errDetails := d.parseData(opts)
		if errDetails.Message != "" {
			return errDetails
		}

		return core.ErrorDetails{}
	}

	d.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}

func (d *DateValidator) IsAfter(opts ValidateOption, after time.Time) {
	vf := func(opts ValidateOption) core.ErrorDetails {
		date, errDetails := d.parseData(opts)
		if errDetails.Message != "" {
			return errDetails
		}

		if date.Before(after) {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	d.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}

func (d *DateValidator) IsBefore(opts ValidateOption, before time.Time) {
	vf := func(opts ValidateOption) core.ErrorDetails {
		date, errDetails := d.parseData(opts)
		if errDetails.Message != "" {
			return errDetails
		}

		if date.After(before) {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	d.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}

func (d *DateValidator) IsBetween(opts ValidateOption, after, before time.Time) {

	vf := func(opts ValidateOption) core.ErrorDetails {
		date, errDetails := d.parseData(opts)
		if errDetails.Message != "" {
			return errDetails
		}

		if date.Before(after) || date.After(before) {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	d.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}

func (d *DateValidator) parseData(opts ValidateOption) (time.Time, core.ErrorDetails) {
	switch v := opts.Data.(type) {
	case time.Time:
		return v, core.ErrorDetails{}
	case string:
		date, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, opts.ToInvalidDataType()
		}
		return date, core.ErrorDetails{}
	case *time.Time:
		if v == nil {
			return time.Time{}, opts.ToInvalidDataType()
		}
		return *v, core.ErrorDetails{}
	case *string:
		if v == nil {
			return time.Time{}, opts.ToInvalidDataType()
		}
		date, err := time.Parse(time.RFC3339, *v)
		if err != nil {
			return time.Time{}, opts.ToInvalidDataType()
		}
		return date, core.ErrorDetails{}
	default:
		return time.Time{}, opts.ToInvalidDataType()
	}
}

func (d *DateValidator) Validate() error {
	return d.Validator.Validate()
}
