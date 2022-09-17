package baseError

import (
	"encoding/json"
	"net/http"
)

func (e BaseError) GenerateResponse() []byte {
	errorResponse, err := json.Marshal(e)
	if err != nil {
		internalServerError, _ := json.Marshal(NewInternalServerError())
		return internalServerError
	}
	return errorResponse
}

func NewInternalServerError() error {
	return &BaseError{
		Message: "internal server error",
		Status:  http.StatusInternalServerError,
	}
}
func NewUnAuthorizedError() error {
	return &BaseError{
		Message: "Unauthorized",
		Status:  http.StatusUnauthorized,
	}
}

func NewBadRequest(message string) error {
	return &BaseError{
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

func NewForbiddenRequest(message string) error {
	return &BaseError{
		Message: message,
		Status:  http.StatusForbidden,
	}
}

func NewValidationError(validationErrors ...ValidationFieldError) error {
	return &BaseError{
		Fields:  validationErrors,
		Message: "invalid data",
		Status:  http.StatusForbidden,
	}
}

func NewFieldValidationError(field string, message string) ValidationFieldError {
	return ValidationFieldError{
		Field:   field,
		Message: message,
	}
}
