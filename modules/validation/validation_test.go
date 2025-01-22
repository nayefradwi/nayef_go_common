package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringValidatorWithPointer(t *testing.T) {
	//Arrange
	email := "test@example.com"
	strValidator := NewStringValidator()

	//Act
	strValidator.IsEmail(ValidateOptionsFrom(email, "email", "Email is required"))
	strValidator.IsEmail(ValidateOptionsFrom(&email, "email", "Email is required"))

	//Assert
	err := strValidator.Validate()
	assert.NoError(t, err)
}
