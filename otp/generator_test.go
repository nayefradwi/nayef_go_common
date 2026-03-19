package otp

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateOtp_NumericLength(t *testing.T) {
	gen := NewCodeGenerator(6, false)
	code := gen.GenerateOtp()
	require.Len(t, code, 6)
	for _, c := range code {
		assert.True(t, unicode.IsDigit(c), "expected digit, got %c", c)
	}
}

func TestGenerateOtp_AlphaNumericLength(t *testing.T) {
	gen := NewCodeGenerator(8, true)
	code := gen.GenerateOtp()
	require.Len(t, code, 8)
	for _, c := range code {
		assert.True(t, unicode.IsLetter(c) || unicode.IsDigit(c), "expected alphanumeric, got %c", c)
	}
}

func TestGenerateOtp_MultipleCodesAreDifferent(t *testing.T) {
	gen := NewCodeGenerator(6, false)
	codes := make(map[string]struct{})
	for range 10 {
		codes[gen.GenerateOtp()] = struct{}{}
	}
	assert.Greater(t, len(codes), 1, "expected multiple distinct codes from 10 generations")
}

func TestHashCode_Deterministic(t *testing.T) {
	hash1 := HashCode("123456")
	hash2 := HashCode("123456")
	assert.Equal(t, hash1, hash2)
}

func TestHashCode_DifferentInputs(t *testing.T) {
	hash1 := HashCode("123456")
	hash2 := HashCode("654321")
	assert.NotEqual(t, hash1, hash2)
}
