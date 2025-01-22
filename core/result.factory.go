package core

func BadRequestError(message string) *ResultError {
	return NewResultError(message, BAD_REQUEST_CODE)
}

func UnauthorizedError(message string) *ResultError {
	return NewResultError(message, UNAUTHORIZED_CODE)
}

func ForbiddenError(message string) *ResultError {
	return NewResultError(message, FORBIDDEN_CODE)
}

func NotFoundError(message string) *ResultError {
	return NewResultError(message, NOT_FOUND_CODE)
}

func InternalError(message string) *ResultError {
	return NewResultError(message, INTERNAL_ERROR_CODE)
}

func InvalidInputError(message string) *ResultError {
	return NewResultError(message, INVALID_INPUT_CODE)
}

func UnknownError(message string) *ResultError {
	return NewResultError(message, UNKNOWN_ERROR_CODE)
}

func NewValidationError(errors []ErrorDetails) *ResultError {
	return NewResultError("Invalid", VALIDATION_ERROR_CODE, errors...)
}
