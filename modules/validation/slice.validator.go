package validation

import (
	"fmt"

	"github.com/nayefradwi/nayef_go_common/result"
)

type SliceValidationRuleFactory[E any] struct{}

func NewSliceValidationRuleFactory[E any]() SliceValidationRuleFactory[E] {
	return SliceValidationRuleFactory[E]{}
}

func (f SliceValidationRuleFactory[E]) Must(
	data []E,
	field string,
	message string,
	ruleCb func(opts ValidationRuleOption[[]E]) result.ErrorDetails,
) ValidationRule[[]E] {
	return ValidationRule[[]E]{
		Validate: ruleCb,
		Opts: ValidationRuleOption[[]E]{
			Field:   field,
			Message: message,
			Data:    data,
		},
	}
}

func (f SliceValidationRuleFactory[E]) NotNilOrEmpty(data []E, field string) ValidationRule[[]E] {
	return ValidationRule[[]E]{
		Opts: ValidationRuleOption[[]E]{
			Field:   field,
			Message: fmt.Sprintf("%s cannot be nil or empty", field),
			Data:    data,
		},
		Validate: func(opts ValidationRuleOption[[]E]) result.ErrorDetails {
			if data == nil {
				return result.ErrorDetails{
					Field:   opts.Field,
					Message: opts.Message,
					Code:    result.INVALID_INPUT_CODE,
				}
			}

			if len(data) == 0 {
				return result.ErrorDetails{
					Field:   opts.Field,
					Message: opts.Message,
					Code:    result.INVALID_INPUT_CODE,
				}
			}

			return result.ErrorDetails{}
		},
	}
}
