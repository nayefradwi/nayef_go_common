package errors

import "net/http"

const (
	CodeNotFound     = "NOT_FOUND"
	CodeBadRequest   = "BAD_REQUEST"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
	CodeInternal     = "INTERNAL_ERROR"
	CodeInvalidInput = "INVALID_INPUT"
	CodeUnknown      = "UNKNOWN_ERROR"
	CodeValidation   = "VALIDATION_ERROR"
)

func BadRequestError(message string) *ResultError {
	return NewResultErrorWithStatus(message, CodeBadRequest, http.StatusBadRequest)
}

func UnauthorizedError(message string) *ResultError {
	return NewResultErrorWithStatus(message, CodeUnauthorized, http.StatusUnauthorized)
}

func ForbiddenError(message string) *ResultError {
	return NewResultErrorWithStatus(message, CodeForbidden, http.StatusForbidden)
}

func NotFoundError(message string) *ResultError {
	return NewResultErrorWithStatus(message, CodeNotFound, http.StatusNotFound)
}

func InternalError(message string) *ResultError {
	return NewResultErrorWithStatus(message, CodeInternal, http.StatusInternalServerError)
}

func InvalidInputError(message string) *ResultError {
	return NewResultErrorWithStatus(message, CodeInvalidInput, http.StatusUnprocessableEntity)
}

func UnknownError(message string) *ResultError {
	return NewResultErrorWithStatus(message, CodeUnknown, http.StatusInternalServerError)
}

func NewValidationError(details ...ErrorDetails) *ResultError {
	return NewResultError("Invalid", CodeValidation, details...)
}

func Field(field, code, message string) ErrorDetails {
	return ErrorDetails{
		Field:   field,
		Code:    code,
		Message: message,
	}
}
