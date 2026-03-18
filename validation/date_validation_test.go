package validation

import (
	"testing"
	"time"

	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/stretchr/testify/assert"
)

func TestIsDate_ZeroTime(t *testing.T) {
	factory := NewDateValidationRuleFactory()
	rule := factory.IsDate(time.Time{}, "created_at")
	err := rule.Validate(rule.Opts)
	assert.NotEmpty(t, err.Message)
	assert.Equal(t, CodeInvalidInput, err.Code)
}

func TestIsDate_ValidTime(t *testing.T) {
	factory := NewDateValidationRuleFactory()
	rule := factory.IsDate(time.Now(), "created_at")
	assert.Empty(t, rule.Validate(rule.Opts).Message)
}

func TestIsAfter_DateBefore(t *testing.T) {
	factory := NewDateValidationRuleFactory()
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	rule := factory.IsAfter(past, "start_date", now)
	err := rule.Validate(rule.Opts)
	assert.NotEmpty(t, err.Message)
	assert.Equal(t, CodeInvalidInput, err.Code)
}

// IsAfter uses data.Before(after) — equal is not "before", so equal passes.
func TestIsAfter_DateEqual(t *testing.T) {
	factory := NewDateValidationRuleFactory()
	now := time.Now().Truncate(time.Second)
	rule := factory.IsAfter(now, "start_date", now)
	assert.Empty(t, rule.Validate(rule.Opts).Message)
}

func TestIsAfter_DateAfter(t *testing.T) {
	factory := NewDateValidationRuleFactory()
	now := time.Now()
	future := now.Add(24 * time.Hour)
	rule := factory.IsAfter(future, "start_date", now)
	assert.Empty(t, rule.Validate(rule.Opts).Message)
}

func TestIsBefore_DateAfter(t *testing.T) {
	factory := NewDateValidationRuleFactory()
	now := time.Now()
	future := now.Add(24 * time.Hour)
	rule := factory.IsBefore(future, "end_date", now)
	err := rule.Validate(rule.Opts)
	assert.NotEmpty(t, err.Message)
	assert.Equal(t, CodeInvalidInput, err.Code)
}

// IsBefore uses data.After(before) — equal is not "after", so equal passes.
func TestIsBefore_DateEqual(t *testing.T) {
	factory := NewDateValidationRuleFactory()
	now := time.Now().Truncate(time.Second)
	rule := factory.IsBefore(now, "end_date", now)
	assert.Empty(t, rule.Validate(rule.Opts).Message)
}

func TestIsBefore_DateBefore(t *testing.T) {
	factory := NewDateValidationRuleFactory()
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	rule := factory.IsBefore(past, "end_date", now)
	assert.Empty(t, rule.Validate(rule.Opts).Message)
}

func TestIsBetween_OutsideRange(t *testing.T) {
	factory := NewDateValidationRuleFactory()
	now := time.Now()
	start := now.Add(1 * time.Hour)
	end := now.Add(2 * time.Hour)
	rule := factory.IsBetween(now, "event_date", start, end)
	err := rule.Validate(rule.Opts)
	assert.NotEmpty(t, err.Message)
	assert.Equal(t, CodeInvalidInput, err.Code)
}

// IsBetween uses Before(start) || After(end) — equal to start passes (inclusive).
func TestIsBetween_BoundaryStart(t *testing.T) {
	factory := NewDateValidationRuleFactory()
	start := time.Now().Truncate(time.Second)
	end := start.Add(2 * time.Hour)
	rule := factory.IsBetween(start, "event_date", start, end)
	assert.Empty(t, rule.Validate(rule.Opts).Message)
}

// Equal to end passes (inclusive).
func TestIsBetween_BoundaryEnd(t *testing.T) {
	factory := NewDateValidationRuleFactory()
	start := time.Now().Truncate(time.Second)
	end := start.Add(2 * time.Hour)
	rule := factory.IsBetween(end, "event_date", start, end)
	assert.Empty(t, rule.Validate(rule.Opts).Message)
}

func TestIsBetween_InRange(t *testing.T) {
	factory := NewDateValidationRuleFactory()
	start := time.Now()
	mid := start.Add(1 * time.Hour)
	end := start.Add(2 * time.Hour)
	rule := factory.IsBetween(mid, "event_date", start, end)
	assert.Empty(t, rule.Validate(rule.Opts).Message)
}

func TestDateMust_Pass(t *testing.T) {
	factory := NewDateValidationRuleFactory()
	now := time.Now()
	rule := factory.Must(now, "event_date", "must not be zero", func(opts ValidationRuleOption[time.Time]) ErrorDetails {
		return ErrorDetails{}
	})
	assert.Empty(t, rule.Validate(rule.Opts).Message)
}

func TestDateMust_Fail(t *testing.T) {
	factory := NewDateValidationRuleFactory()
	past := time.Now().Add(-48 * time.Hour)
	cutoff := time.Now().Add(-24 * time.Hour)
	rule := factory.Must(past, "event_date", "must be within last 24 hours", func(opts ValidationRuleOption[time.Time]) ErrorDetails {
		if opts.Data.Before(cutoff) {
			return ErrorDetails{Field: opts.Field, Message: opts.Message, Code: CodeInvalidInput}
		}
		return ErrorDetails{}
	})
	err := rule.Validate(rule.Opts)
	assert.NotEmpty(t, err.Message)
	assert.Equal(t, CodeInvalidInput, err.Code)
}
