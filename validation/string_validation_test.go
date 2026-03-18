package validation

import (
	"testing"

	. "github.com/nayefradwi/nayef_go_common/errors"
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
	assert.Equal(t, CodeInvalidInput, err.Code)

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

	// Passing cases
	assert.Empty(t, factory.MinLength("hello", "name", 3).Validate(factory.MinLength("hello", "name", 3).Opts).Message)
	assert.Empty(t, factory.MaxLength("hi", "name", 5).Validate(factory.MaxLength("hi", "name", 5).Opts).Message)
	assert.Empty(t, factory.ExactLength("abc", "code", 3).Validate(factory.ExactLength("abc", "code", 3).Opts).Message)

	// Boundary cases
	assert.Empty(t, factory.MinLength("abc", "name", 3).Validate(factory.MinLength("abc", "name", 3).Opts).Message)
	assert.Empty(t, factory.MaxLength("abcde", "name", 5).Validate(factory.MaxLength("abcde", "name", 5).Opts).Message)
}

func TestStringValidation_Patterns(t *testing.T) {
	factory := NewStringValidationRuleFactory()

	alpha := factory.IsAlpha("abcXYZ", "name")
	numeric := factory.IsNumeric("12345", "id")
	alphanumeric := factory.IsAlphanumeric("user123", "username")

	assert.Empty(t, alpha.Validate(alpha.Opts).Message)
	assert.Empty(t, numeric.Validate(numeric.Opts).Message)
	assert.Empty(t, alphanumeric.Validate(alphanumeric.Opts).Message)

	// Invalid cases
	invalidAlpha := factory.IsAlpha("abc123", "name")
	invalidNumeric := factory.IsNumeric("123abc", "id")
	invalidAlphanumeric := factory.IsAlphanumeric("user@name", "username")

	assert.NotEmpty(t, invalidAlpha.Validate(invalidAlpha.Opts).Message)
	assert.NotEmpty(t, invalidNumeric.Validate(invalidNumeric.Opts).Message)
	assert.NotEmpty(t, invalidAlphanumeric.Validate(invalidAlphanumeric.Opts).Message)
}

func TestStringValidation_MatchesPattern(t *testing.T) {
	factory := NewStringValidationRuleFactory()

	valid := factory.MatchesPattern("hello-world", "slug", `^[a-z-]+$`)
	invalid := factory.MatchesPattern("Hello World", "slug", `^[a-z-]+$`)
	badPattern := factory.MatchesPattern("hello", "slug", `[invalid`)

	assert.Empty(t, valid.Validate(valid.Opts).Message)
	assert.NotEmpty(t, invalid.Validate(invalid.Opts).Message)
	assert.Equal(t, CodeInvalidInput, invalid.Validate(invalid.Opts).Code)
	assert.NotEmpty(t, badPattern.Validate(badPattern.Opts).Message)
}

func TestStringValidation_Must(t *testing.T) {
	factory := NewStringValidationRuleFactory()

	noSpaces := func(opts ValidationRuleOption[string]) ErrorDetails {
		for _, c := range opts.Data {
			if c == ' ' {
				return ErrorDetails{Field: opts.Field, Message: opts.Message, Code: CodeInvalidInput}
			}
		}
		return ErrorDetails{}
	}

	valid := factory.Must("hello", "slug", "must not contain spaces", noSpaces)
	invalid := factory.Must("hello world", "slug", "must not contain spaces", noSpaces)

	assert.Empty(t, valid.Validate(valid.Opts).Message)
	assert.NotEmpty(t, invalid.Validate(invalid.Opts).Message)
	assert.Equal(t, CodeInvalidInput, invalid.Validate(invalid.Opts).Code)
}

func TestStringValidation_ErrorMessageContainsFieldName(t *testing.T) {
	factory := NewStringValidationRuleFactory()

	rule := factory.IsRequired("", "username")
	err := rule.Validate(rule.Opts)
	assert.Contains(t, err.Message, "username")

	minRule := factory.MinLength("hi", "username", 5)
	minErr := minRule.Validate(minRule.Opts)
	assert.Contains(t, minErr.Message, "username")
}
