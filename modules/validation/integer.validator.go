package validation

import (
	"fmt"

	"github.com/nayefradwi/nayef_go_common/core"
)

type NumValidationRuleFactory[T int64 | int32 | int16 | int8 | int | float64 | float32] struct{}

func NewNumValidationRuleFactory[T int64 | int32 | int16 | int8 | int | float64 | float32]() NumValidationRuleFactory[T] {
	return NumValidationRuleFactory[T]{}
}

func (n NumValidationRuleFactory[T]) Must(data T, field string, message string, ruleCb func(opts ValidationRuleOption[T]) core.ErrorDetails) ValidationRule[T] {
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
		Validate: func(opts ValidationRuleOption[T]) core.ErrorDetails {
			if data < min {
				return core.ErrorDetails{
					Message: opts.Message,
					Code:    core.INVALID_INPUT_CODE,
					Field:   opts.Field,
				}
			}

			return core.ErrorDetails{}
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
		Validate: func(opts ValidationRuleOption[T]) core.ErrorDetails {
			if data > max {
				return core.ErrorDetails{
					Message: opts.Message,
					Code:    core.INVALID_INPUT_CODE,
					Field:   opts.Field,
				}
			}
			return core.ErrorDetails{}
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
		Validate: func(opts ValidationRuleOption[T]) core.ErrorDetails {
			if data < min || data > max {
				return core.ErrorDetails{
					Message: opts.Message,
					Code:    core.INVALID_INPUT_CODE,
					Field:   opts.Field,
				}
			}
			return core.ErrorDetails{}
		},
	}
}
