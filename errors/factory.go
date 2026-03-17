package errors

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
	return NewResultError(message, CodeBadRequest)
}

func UnauthorizedError(message string) *ResultError {
	return NewResultError(message, CodeUnauthorized)
}

func ForbiddenError(message string) *ResultError {
	return NewResultError(message, CodeForbidden)
}

func NotFoundError(message string) *ResultError {
	return NewResultError(message, CodeNotFound)
}

func InternalError(message string) *ResultError {
	return NewResultError(message, CodeInternal)
}

func InvalidInputError(message string) *ResultError {
	return NewResultError(message, CodeInvalidInput)
}

func UnknownError(message string) *ResultError {
	return NewResultError(message, CodeUnknown)
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
