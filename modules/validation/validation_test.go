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

func TestStringValidatorWithInt(t *testing.T) {
	//Arrange
	age := 20
	strValidator := NewStringValidator()

	//Act
	strValidator.IsEmail(ValidateOptionsFrom(age, "email", "Email is required"))

	//Assert
	err := strValidator.Validate()
	assert.Error(t, err)
}

func TestStringValidatorWithNil(t *testing.T) {
	//Arrange
	strValidator := NewStringValidator()

	//Act
	strValidator.IsEmail(ValidateOptionsFrom(nil, "email", "Email is required"))

	//Assert
	err := strValidator.Validate()
	assert.Error(t, err)
}
