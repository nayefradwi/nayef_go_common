package validation

import (
	"fmt"
	"regexp"
	"unicode/utf8"

	. "github.com/nayefradwi/nayef_go_common/errors"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type StringValidationRuleFactory struct{}

func NewStringValidationRuleFactory() StringValidationRuleFactory {
	return StringValidationRuleFactory{}
}

func (f StringValidationRuleFactory) Must(
	data string,
	field string,
	message string,
	ruleCb func(opts ValidationRuleOption[string]) ErrorDetails,
) ValidationRule[string] {
	return ValidationRule[string]{
		Validate: ruleCb,
		Opts: ValidationRuleOption[string]{
			Field:   field,
			Message: message,
			Data:    data,
		},
	}
}

func (f StringValidationRuleFactory) IsRequired(data string, field string) ValidationRule[string] {
	return ValidationRule[string]{
		Opts: ValidationRuleOption[string]{
			Field:   field,
			Message: fmt.Sprintf("%s is required", field),
			Data:    data,
		},
		Validate: func(opts ValidationRuleOption[string]) ErrorDetails {
			if opts.Data == "" {
				return ErrorDetails{
					Field:   opts.Field,
					Message: opts.Message,
					Code:    CodeInvalidInput,
				}
			}
			return ErrorDetails{}
		},
	}
}

func (f StringValidationRuleFactory) MinLength(data string, field string, min int) ValidationRule[string] {
	return ValidationRule[string]{
		Opts: ValidationRuleOption[string]{
			Field:   field,
			Message: fmt.Sprintf("%s must be at least %d characters long", field, min),
			Data:    data,
		},
		Validate: func(opts ValidationRuleOption[string]) ErrorDetails {
			if utf8.RuneCountInString(opts.Data) < min {
				return ErrorDetails{
					Field:   opts.Field,
					Message: opts.Message,
					Code:    CodeInvalidInput,
				}
			}
			return ErrorDetails{}
		},
	}
}

func (f StringValidationRuleFactory) MaxLength(data string, field string, max int) ValidationRule[string] {
	return ValidationRule[string]{
		Opts: ValidationRuleOption[string]{
			Field:   field,
			Message: fmt.Sprintf("%s cannot exceed %d characters", field, max),
			Data:    data,
		},
		Validate: func(opts ValidationRuleOption[string]) ErrorDetails {
			if utf8.RuneCountInString(opts.Data) > max {
				return ErrorDetails{
					Field:   opts.Field,
					Message: opts.Message,
					Code:    CodeInvalidInput,
				}
			}
			return ErrorDetails{}
		},
	}
}

func (f StringValidationRuleFactory) ExactLength(data string, field string, length int) ValidationRule[string] {
	return ValidationRule[string]{
		Opts: ValidationRuleOption[string]{
			Field:   field,
			Message: fmt.Sprintf("%s must be exactly %d characters long", field, length),
			Data:    data,
		},
		Validate: func(opts ValidationRuleOption[string]) ErrorDetails {
			if utf8.RuneCountInString(opts.Data) != length {
				return ErrorDetails{
					Field:   opts.Field,
					Message: opts.Message,
					Code:    CodeInvalidInput,
				}
			}
			return ErrorDetails{}
		},
	}
}

func (f StringValidationRuleFactory) MatchesPattern(data string, field, pattern string) ValidationRule[string] {
	re, compileErr := regexp.Compile(pattern)
	return ValidationRule[string]{
		Opts: ValidationRuleOption[string]{
			Field:   field,
			Message: fmt.Sprintf("%s must match pattern %q", field, pattern),
			Data:    data,
		},
		Validate: func(opts ValidationRuleOption[string]) ErrorDetails {
			if compileErr != nil || !re.MatchString(opts.Data) {
				return ErrorDetails{
					Field:   opts.Field,
					Message: opts.Message,
					Code:    CodeInvalidInput,
				}
			}
			return ErrorDetails{}
		},
	}
}

func (f StringValidationRuleFactory) IsEmail(data string, field string) ValidationRule[string] {
	return ValidationRule[string]{
		Opts: ValidationRuleOption[string]{
			Field:   field,
			Message: fmt.Sprintf("%s must be a valid email address", field),
			Data:    data,
		},
		Validate: func(opts ValidationRuleOption[string]) ErrorDetails {
			if !emailRegex.MatchString(opts.Data) {
				return ErrorDetails{
					Field:   opts.Field,
					Message: opts.Message,
					Code:    CodeInvalidInput,
				}
			}
			return ErrorDetails{}
		},
	}
}

func (f StringValidationRuleFactory) IsAlpha(data string, field string) ValidationRule[string] {
	return f.MatchesPattern(data, field, `^[a-zA-Z]*$`)
}

func (f StringValidationRuleFactory) IsNumeric(data string, field string) ValidationRule[string] {
	return f.MatchesPattern(data, field, `^[0-9]*$`)
}

func (f StringValidationRuleFactory) IsAlphanumeric(data string, field string) ValidationRule[string] {
	return f.MatchesPattern(data, field, `^[a-zA-Z0-9]*$`)
}
