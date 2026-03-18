package validation

import . "github.com/nayefradwi/nayef_go_common/errors"

type IValidator interface {
	Validate() error
}

type IValidationRuleFactory[T any] interface {
	Must(data T, field, message string, ruleCb func(opts ValidationRuleOption[T]) ErrorDetails) ValidationRule[T]
}

type ValidationRuleOption[T any] struct {
	Field   string
	Message string
	Data    T
}

type ValidationRule[T any] struct {
	Opts     ValidationRuleOption[T]
	Validate func(opts ValidationRuleOption[T]) ErrorDetails
}

type Validator struct {
	Rules []ValidationRule[any]
}

func NewValidator() *Validator {
	validator := &Validator{}
	validator.Rules = make([]ValidationRule[any], 0)
	return validator
}

func AddRule[T any](v *Validator, rule ValidationRule[T]) {
	v.Rules = append(v.Rules, toAny(rule))
}

func toAny[T any](rule ValidationRule[T]) ValidationRule[any] {
	return ValidationRule[any]{
		Opts: ValidationRuleOption[any]{
			Field:   rule.Opts.Field,
			Message: rule.Opts.Message,
			Data:    rule.Opts.Data,
		},
		Validate: func(opts ValidationRuleOption[any]) ErrorDetails {
			typedOpts := ValidationRuleOption[T]{
				Field:   opts.Field,
				Message: opts.Message,
				Data:    opts.Data.(T),
			}
			return rule.Validate(typedOpts)
		},
	}
}

func (v *Validator) Validate() error {
	errorDetails := make([]ErrorDetails, 0)
	hasError := false

	for _, rule := range v.Rules {
		errDetails := rule.Validate(rule.Opts)
		if errDetails.Message != "" {
			hasError = true
			errorDetails = append(errorDetails, errDetails)
		}
	}

	if hasError {
		return NewValidationError(errorDetails...)
	}

	return nil
}
