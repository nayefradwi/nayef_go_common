package validation

import (
	"fmt"

	. "github.com/nayefradwi/nayef_go_common/errors"
)

type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

type NumValidationRuleFactory[T Numeric] struct{}

func NewNumValidationRuleFactory[T Numeric]() NumValidationRuleFactory[T] {
	return NumValidationRuleFactory[T]{}
}

func (n NumValidationRuleFactory[T]) Must(data T, field string, message string, ruleCb func(opts ValidationRuleOption[T]) ErrorDetails) ValidationRule[T] {
	return ValidationRule[T]{
		Validate: ruleCb,
		Opts: ValidationRuleOption[T]{
			Field:   field,
			Message: message,
			Data:    data,
		},
	}
}

func (f NumValidationRuleFactory[T]) MinValue(data T, field string, min T) ValidationRule[T] {
	return ValidationRule[T]{
		Opts: ValidationRuleOption[T]{
			Field:   field,
			Message: fmt.Sprintf("%s cannot be less than %v", field, min),
			Data:    data,
		},
		Validate: func(opts ValidationRuleOption[T]) ErrorDetails {
			if opts.Data < min {
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

func (f NumValidationRuleFactory[T]) MaxValue(data T, field string, max T) ValidationRule[T] {
	return ValidationRule[T]{
		Opts: ValidationRuleOption[T]{
			Field:   field,
			Message: fmt.Sprintf("%s cannot be greater than %v", field, max),
			Data:    data,
		},
		Validate: func(opts ValidationRuleOption[T]) ErrorDetails {
			if opts.Data > max {
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

func (f NumValidationRuleFactory[T]) Between(data T, field string, min, max T) ValidationRule[T] {
	return ValidationRule[T]{
		Opts: ValidationRuleOption[T]{
			Field:   field,
			Message: fmt.Sprintf("%s must be between %v and %v", field, min, max),
			Data:    data,
		},
		Validate: func(opts ValidationRuleOption[T]) ErrorDetails {
			if opts.Data < min || opts.Data > max {
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
