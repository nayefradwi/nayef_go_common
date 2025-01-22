package validation

import (
	"regexp"

	"github.com/nayefradwi/nayef_go_common/core"
)

type StringValidator struct {
	*Validator
}

func NewStringValidator() *StringValidator {
	return &StringValidator{Validator: NewValidator()}
}

func StringValidatorFromValidator(validator *Validator) *StringValidator {
	return &StringValidator{Validator: validator}
}

func (s *StringValidator) MatchesPattern(opts ValidateOption, pattern string) {
	vf := func(opts ValidateOption) core.ErrorDetails {
		str, ok := opts.Data.(string)
		if !ok {
			return core.ErrorDetails{Field: opts.Field, Message: "Invalid data type", Code: INVALID_DATA_TYPE}
		}

		matched, err := regexp.MatchString(pattern, str)
		if err != nil || !matched {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	s.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}

func (s *StringValidator) MaxLength(opts ValidateOption, max int) {
	vf := func(opts ValidateOption) core.ErrorDetails {
		str, ok := opts.Data.(string)
		if !ok {
			return core.ErrorDetails{Field: opts.Field, Message: "Invalid data type", Code: INVALID_DATA_TYPE}
		}

		if len(str) > max {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	s.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}

func (s *StringValidator) MinLength(opts ValidateOption, min int) {
	vf := func(opts ValidateOption) core.ErrorDetails {
		str, ok := opts.Data.(string)
		if !ok {
			return core.ErrorDetails{Field: opts.Field, Message: "Invalid data type", Code: INVALID_DATA_TYPE}
		}

		if len(str) < min {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	s.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}

func (s *StringValidator) ExactLength(opts ValidateOption, length int) {
	vf := func(opts ValidateOption) core.ErrorDetails {
		str, ok := opts.Data.(string)
		if !ok {
			return core.ErrorDetails{Field: opts.Field, Message: "Invalid data type", Code: INVALID_DATA_TYPE}
		}

		if len(str) != length {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	s.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}

func (s *StringValidator) IsEmail(opts ValidateOption) {
	vf := func(opts ValidateOption) core.ErrorDetails {
		str, ok := opts.Data.(string)
		if !ok {
			return core.ErrorDetails{Field: opts.Field, Message: "Invalid data type", Code: INVALID_DATA_TYPE}
		}

		emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		matched, err := regexp.MatchString(emailRegex, str)
		if err != nil || !matched {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	s.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}

func (s *StringValidator) IsRequired(opts ValidateOption) {
	vf := func(opts ValidateOption) core.ErrorDetails {
		str, ok := opts.Data.(string)
		if !ok {
			return core.ErrorDetails{Field: opts.Field, Message: "Invalid data type", Code: INVALID_DATA_TYPE}
		}

		if str == "" {
			return core.ErrorDetails{Field: opts.Field, Message: opts.Message}
		}

		return core.ErrorDetails{}
	}

	s.Validator.AddValidation(ValidationFunc{Opts: opts, fn: vf})
}

func (s *StringValidator) Validate() error {
	return s.Validator.Validate()
}
