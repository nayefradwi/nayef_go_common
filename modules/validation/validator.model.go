package validation

import "github.com/nayefradwi/nayef_go_common/core"

const (
	INVALID_DATA_TYPE = "INVALID_DATA_TYPE"
)

type IValidator interface {
	Validate() error
}

type IValidationRuleFactory[T any] interface {
	Must(data T, field, message string, ruleCb func(opts ValidationRuleOption[T]) core.ErrorDetails) ValidationRule[T]
}

type ValidationRuleOption[T any] struct {
	Field   string
	Message string
	Data    T
}

type ValidationRule[T any] struct {
	Opts     ValidationRuleOption[T]
	Validate func(opts ValidationRuleOption[T]) core.ErrorDetails
}

type Validator struct {
	Rules []ValidationRule[any]
}

func NewValidator() *Validator {
	validator := &Validator{}
	validator.Rules = make([]ValidationRule[any], 0)
	return validator
}

func (v *Validator) AddValidation(fn ValidationRule[any]) {
	v.Rules = append(v.Rules, fn)
}

func (v *Validator) Validate() error {
	errorDetails := make([]core.ErrorDetails, 0)
	hasError := false

	for _, rule := range v.Rules {
		errDetails := rule.Validate(rule.Opts)
		if errDetails.Message != "" {
			hasError = true
			errorDetails = append(errorDetails, errDetails)
		}
	}

	if hasError {
		return core.NewValidationError(errorDetails)
	}

	return nil
}
