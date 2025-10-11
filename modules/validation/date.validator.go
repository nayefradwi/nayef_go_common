package validation

import (
	"fmt"
	"time"

	"github.com/nayefradwi/nayef_go_common/result"
)

type DateValidationRuleFactory struct{}

func NewDateValidationRuleFactory() DateValidationRuleFactory {
	return DateValidationRuleFactory{}
}

func (f DateValidationRuleFactory) Must(
	data time.Time,
	field string,
	message string,
	ruleCb func(opts ValidationRuleOption[time.Time]) result.ErrorDetails,
) ValidationRule[time.Time] {
	return ValidationRule[time.Time]{
		Validate: ruleCb,
		Opts: ValidationRuleOption[time.Time]{
			Field:   field,
			Message: message,
			Data:    data,
		},
	}
}

func (f DateValidationRuleFactory) IsDate(data time.Time, field string) ValidationRule[time.Time] {
	return ValidationRule[time.Time]{
		Opts: ValidationRuleOption[time.Time]{
			Field:   field,
			Message: fmt.Sprintf("%s must be a valid date", field),
			Data:    data,
		},
		Validate: func(opts ValidationRuleOption[time.Time]) result.ErrorDetails {
			if data.IsZero() {
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

func (f DateValidationRuleFactory) IsAfter(data time.Time, field string, after time.Time) ValidationRule[time.Time] {
	return ValidationRule[time.Time]{
		Opts: ValidationRuleOption[time.Time]{
			Field:   field,
			Message: fmt.Sprintf("%s must be after %s", field, after.Format(time.RFC3339)),
			Data:    data,
		},
		Validate: func(opts ValidationRuleOption[time.Time]) result.ErrorDetails {
			if data.Before(after) {
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

func (f DateValidationRuleFactory) IsBefore(data time.Time, field string, before time.Time) ValidationRule[time.Time] {
	return ValidationRule[time.Time]{
		Opts: ValidationRuleOption[time.Time]{
			Field:   field,
			Message: fmt.Sprintf("%s must be before %s", field, before.Format(time.RFC3339)),
			Data:    data,
		},
		Validate: func(opts ValidationRuleOption[time.Time]) result.ErrorDetails {
			if data.After(before) {
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

func (f DateValidationRuleFactory) IsBetween(data time.Time, field string, start, end time.Time) ValidationRule[time.Time] {
	return ValidationRule[time.Time]{
		Opts: ValidationRuleOption[time.Time]{
			Field:   field,
			Message: fmt.Sprintf("%s must be between %s and %s", field, start.Format(time.RFC3339), end.Format(time.RFC3339)),
			Data:    data,
		},
		Validate: func(opts ValidationRuleOption[time.Time]) result.ErrorDetails {
			if data.Before(start) || data.After(end) {
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
