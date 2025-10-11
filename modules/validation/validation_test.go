package validation

import (
	"testing"
	"time"

	"github.com/nayefradwi/nayef_go_common/core"
	"github.com/stretchr/testify/assert"
)

//
// ---------- STRING VALIDATION FACTORY TESTS ----------
//

func TestStringValidation_IsEmail_ValidAndInvalid(t *testing.T) {
	factory := NewStringValidationRuleFactory()

	validRule := factory.IsEmail("test@example.com", "email")
	invalidRule := factory.IsEmail("invalid-email", "email")

	assert.Empty(t, validRule.Validate(validRule.Opts).Message)
	assert.NotEmpty(t, invalidRule.Validate(invalidRule.Opts).Message)
}

func TestStringValidation_IsRequired(t *testing.T) {
	factory := NewStringValidationRuleFactory()

	rule := factory.IsRequired("", "username")
	err := rule.Validate(rule.Opts)

	assert.NotEmpty(t, err.Message)
	assert.Equal(t, core.INVALID_INPUT_CODE, err.Code)

	valid := factory.IsRequired("hello", "username")
	assert.Empty(t, valid.Validate(valid.Opts).Message)
}

func TestStringValidation_LengthChecks(t *testing.T) {
	factory := NewStringValidationRuleFactory()

	minRule := factory.MinLength("hi", "name", 3)
	maxRule := factory.MaxLength("toolongname", "name", 5)
	exactRule := factory.ExactLength("abc", "code", 5)

	assert.NotEmpty(t, minRule.Validate(minRule.Opts).Message)
	assert.NotEmpty(t, maxRule.Validate(maxRule.Opts).Message)
	assert.NotEmpty(t, exactRule.Validate(exactRule.Opts).Message)
}

func TestStringValidation_Patterns(t *testing.T) {
	factory := NewStringValidationRuleFactory()

	alpha := factory.IsAlpha("abcXYZ", "name")
	numeric := factory.IsNumeric("12345", "id")
	alphanumeric := factory.IsAlphanumeric("user123", "username")

	assert.Empty(t, alpha.Validate(alpha.Opts).Message)
	assert.Empty(t, numeric.Validate(numeric.Opts).Message)
	assert.Empty(t, alphanumeric.Validate(alphanumeric.Opts).Message)
}

//
// ---------- NUMERIC VALIDATION FACTORY TESTS ----------
//

func TestNumValidation_MinMaxBetween(t *testing.T) {
	factory := NewNumValidationRuleFactory[int]()

	minRule := factory.MinValue(2, "age", 5)
	maxRule := factory.MaxValue(10, "score", 8)
	betweenRule := factory.Between(15, "points", 5, 10)

	assert.NotEmpty(t, minRule.Validate(minRule.Opts).Message)
	assert.NotEmpty(t, maxRule.Validate(maxRule.Opts).Message)
	assert.NotEmpty(t, betweenRule.Validate(betweenRule.Opts).Message)

	validMin := factory.MinValue(6, "age", 5)
	assert.Empty(t, validMin.Validate(validMin.Opts).Message)
}

//
// ---------- SLICE VALIDATION FACTORY TESTS ----------
//

func TestSliceValidation_NotNilOrEmpty(t *testing.T) {
	factory := NewSliceValidationRuleFactory[int]()

	nilRule := factory.NotNilOrEmpty(nil, "numbers")
	emptyRule := factory.NotNilOrEmpty([]int{}, "numbers")
	validRule := factory.NotNilOrEmpty([]int{1, 2}, "numbers")

	assert.NotEmpty(t, nilRule.Validate(nilRule.Opts).Message)
	assert.NotEmpty(t, emptyRule.Validate(emptyRule.Opts).Message)
	assert.Empty(t, validRule.Validate(validRule.Opts).Message)
}

func TestSliceValidation_CustomMustRule(t *testing.T) {
	factory := NewSliceValidationRuleFactory[string]()

	rule := factory.Must([]string{"a"}, "tags", "must have 2 items", func(opts ValidationRuleOption[[]string]) core.ErrorDetails {
		if len(opts.Data) < 2 {
			return core.ErrorDetails{
				Field:   opts.Field,
				Message: opts.Message,
				Code:    core.INVALID_INPUT_CODE,
			}
		}
		return core.ErrorDetails{}
	})

	assert.NotEmpty(t, rule.Validate(rule.Opts).Message)

	valid := factory.Must([]string{"a", "b"}, "tags", "must have 2 items", func(opts ValidationRuleOption[[]string]) core.ErrorDetails {
		if len(opts.Data) < 2 {
			return core.ErrorDetails{Message: opts.Message}
		}
		return core.ErrorDetails{}
	})

	assert.Empty(t, valid.Validate(valid.Opts).Message)
}

//
// ---------- DATE VALIDATION FACTORY TESTS ----------
//

func TestDateValidation_IsDate_IsAfter_IsBefore_IsBetween(t *testing.T) {
	factory := NewDateValidationRuleFactory()
	now := time.Now()

	// IsDate
	invalidDate := time.Time{}
	validDate := now
	r1 := factory.IsDate(invalidDate, "start_date")
	r2 := factory.IsDate(validDate, "start_date")
	assert.NotEmpty(t, r1.Validate(r1.Opts).Message)
	assert.Empty(t, r2.Validate(r2.Opts).Message)

	// IsAfter
	past := now.Add(-24 * time.Hour)
	r3 := factory.IsAfter(past, "event", now)
	assert.NotEmpty(t, r3.Validate(r3.Opts).Message)
	r4 := factory.IsAfter(now.Add(1*time.Hour), "event", past)
	assert.Empty(t, r4.Validate(r4.Opts).Message)

	// IsBefore
	future := now.Add(24 * time.Hour)
	r5 := factory.IsBefore(future, "event", now)
	assert.NotEmpty(t, r5.Validate(r5.Opts).Message)
	r6 := factory.IsBefore(now.Add(-1*time.Hour), "event", future)
	assert.Empty(t, r6.Validate(r6.Opts).Message)

	// IsBetween
	r7 := factory.IsBetween(now.Add(-2*time.Hour), "meeting", now.Add(-1*time.Hour), now)
	r8 := factory.IsBetween(now, "meeting", now.Add(-1*time.Hour), now.Add(1*time.Hour))

	assert.NotEmpty(t, r7.Validate(r7.Opts).Message)
	assert.Empty(t, r8.Validate(r8.Opts).Message)
}

//
// ---------- GENERIC Must TESTS ----------
//

func TestMust_CustomLogic(t *testing.T) {
	factory := NewNumValidationRuleFactory[int]()

	rule := factory.Must(10, "age", "must be even", func(opts ValidationRuleOption[int]) core.ErrorDetails {
		if opts.Data%2 != 0 {
			return core.ErrorDetails{
				Field:   opts.Field,
				Message: opts.Message,
				Code:    core.INVALID_INPUT_CODE,
			}
		}
		return core.ErrorDetails{}
	})

	assert.Empty(t, rule.Validate(rule.Opts).Message)

	badRule := factory.Must(11, "age", "must be even", func(opts ValidationRuleOption[int]) core.ErrorDetails {
		if opts.Data%2 != 0 {
			return core.ErrorDetails{
				Field:   opts.Field,
				Message: opts.Message,
				Code:    core.INVALID_INPUT_CODE,
			}
		}
		return core.ErrorDetails{}
	})

	assert.NotEmpty(t, badRule.Validate(badRule.Opts).Message)
}
