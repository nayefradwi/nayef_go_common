package validation

import (
	"testing"

	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/stretchr/testify/assert"
)

func TestNotNilOrEmpty_NilSlice(t *testing.T) {
	factory := NewSliceValidationRuleFactory[int]()
	rule := factory.NotNilOrEmpty(nil, "items")
	err := rule.Validate(rule.Opts)
	assert.NotEmpty(t, err.Message)
	assert.Equal(t, CodeInvalidInput, err.Code)
}

func TestNotNilOrEmpty_EmptySlice(t *testing.T) {
	factory := NewSliceValidationRuleFactory[int]()
	rule := factory.NotNilOrEmpty([]int{}, "items")
	err := rule.Validate(rule.Opts)
	assert.NotEmpty(t, err.Message)
	assert.Equal(t, CodeInvalidInput, err.Code)
}

func TestNotNilOrEmpty_ValidSlice(t *testing.T) {
	factory := NewSliceValidationRuleFactory[int]()
	rule := factory.NotNilOrEmpty([]int{1, 2}, "items")
	assert.Empty(t, rule.Validate(rule.Opts).Message)
}

func TestSliceMust_Pass(t *testing.T) {
	factory := NewSliceValidationRuleFactory[int]()
	rule := factory.Must([]int{1, 2, 3}, "items", "must have items", func(opts ValidationRuleOption[[]int]) ErrorDetails {
		return ErrorDetails{}
	})
	assert.Empty(t, rule.Validate(rule.Opts).Message)
}

func TestSliceMust_Fail(t *testing.T) {
	factory := NewSliceValidationRuleFactory[int]()
	rule := factory.Must([]int{1}, "items", "must have at least 2 items", func(opts ValidationRuleOption[[]int]) ErrorDetails {
		if len(opts.Data) < 2 {
			return ErrorDetails{Field: opts.Field, Message: opts.Message, Code: CodeInvalidInput}
		}
		return ErrorDetails{}
	})
	err := rule.Validate(rule.Opts)
	assert.NotEmpty(t, err.Message)
	assert.Equal(t, CodeInvalidInput, err.Code)
}

func TestSliceValidator_Integration(t *testing.T) {
	factory := NewSliceValidationRuleFactory[string]()
	validator := NewValidator()

	nilRule := factory.NotNilOrEmpty(nil, "tags")
	emptyRule := factory.NotNilOrEmpty([]string{}, "labels")
	AddRule(validator, nilRule)
	AddRule(validator, emptyRule)

	err := validator.Validate()
	assert.NotNil(t, err)
	assert.NotEmpty(t, err.Error())
}
