package validation

import (
	"testing"

	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/stretchr/testify/assert"
)

//
// ---------- NUM VALIDATION FACTORY TESTS ----------
//

func TestNumValidation_MinValue(t *testing.T) {
	factory := NewNumValidationRuleFactory[int]()

	fail := factory.MinValue(2, "age", 5)
	err := fail.Validate(fail.Opts)
	assert.NotEmpty(t, err.Message)
	assert.Equal(t, CodeInvalidInput, err.Code)
	assert.Contains(t, err.Message, "age")

	atMin := factory.MinValue(5, "age", 5)
	assert.Empty(t, atMin.Validate(atMin.Opts).Message)

	above := factory.MinValue(10, "age", 5)
	assert.Empty(t, above.Validate(above.Opts).Message)
}

func TestNumValidation_MaxValue(t *testing.T) {
	factory := NewNumValidationRuleFactory[int]()

	fail := factory.MaxValue(10, "score", 8)
	err := fail.Validate(fail.Opts)
	assert.NotEmpty(t, err.Message)
	assert.Equal(t, CodeInvalidInput, err.Code)
	assert.Contains(t, err.Message, "score")

	atMax := factory.MaxValue(8, "score", 8)
	assert.Empty(t, atMax.Validate(atMax.Opts).Message)

	below := factory.MaxValue(5, "score", 8)
	assert.Empty(t, below.Validate(below.Opts).Message)
}

func TestNumValidation_Between(t *testing.T) {
	factory := NewNumValidationRuleFactory[int]()

	failBelow := factory.Between(3, "points", 5, 10)
	assert.NotEmpty(t, failBelow.Validate(failBelow.Opts).Message)

	failAbove := factory.Between(15, "points", 5, 10)
	assert.NotEmpty(t, failAbove.Validate(failAbove.Opts).Message)
	assert.Equal(t, CodeInvalidInput, failAbove.Validate(failAbove.Opts).Code)

	pass := factory.Between(7, "points", 5, 10)
	assert.Empty(t, pass.Validate(pass.Opts).Message)

	atMin := factory.Between(5, "points", 5, 10)
	assert.Empty(t, atMin.Validate(atMin.Opts).Message)

	atMax := factory.Between(10, "points", 5, 10)
	assert.Empty(t, atMax.Validate(atMax.Opts).Message)
}

func TestNumValidation_Must(t *testing.T) {
	factory := NewNumValidationRuleFactory[int]()

	isPositive := func(opts ValidationRuleOption[int]) ErrorDetails {
		if opts.Data <= 0 {
			return ErrorDetails{Field: opts.Field, Message: opts.Message, Code: CodeInvalidInput}
		}
		return ErrorDetails{}
	}

	valid := factory.Must(5, "count", "must be positive", isPositive)
	assert.Empty(t, valid.Validate(valid.Opts).Message)

	invalid := factory.Must(-1, "count", "must be positive", isPositive)
	err := invalid.Validate(invalid.Opts)
	assert.NotEmpty(t, err.Message)
	assert.Equal(t, CodeInvalidInput, err.Code)
}

func TestNumValidation_Float64(t *testing.T) {
	factory := NewNumValidationRuleFactory[float64]()

	fail := factory.MinValue(1.5, "price", 2.0)
	err := fail.Validate(fail.Opts)
	assert.NotEmpty(t, err.Message)
	assert.Equal(t, CodeInvalidInput, err.Code)

	pass := factory.MinValue(2.5, "price", 2.0)
	assert.Empty(t, pass.Validate(pass.Opts).Message)
}
