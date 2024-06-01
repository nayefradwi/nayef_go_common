package common

func newError(status int, code string, message string) *ApiError {
	return &ApiError{
		Message: message,
		Status:  status,
		Code:    code,
	}
}

func NewCustomError(status int, code string, message string, errors ...ErrorDetails) *ApiError {
	return &ApiError{
		Message: message,
		Status:  status,
		Code:    code,
		Errors:  errors,
	}
}

func NewUnAuthorizedError(message string) *ApiError {
	return newError(UNAUTHORIZED, UNAUTHORIZED_CODE, message)
}

func NewInternalServerError() *ApiError {
	return newError(INTERNAL_SERVER_ERROR, INTERNAL_ERROR_CODE, "Internal server error")
}

func NewNotFoundError(message string) *ApiError {
	return newError(NOT_FOUND, NOT_FOUND_CODE, message)
}

func NewBadRequestError(code string, message string) *ApiError {
	return NewCustomError(BAD_REQUEST, code, message)
}

func NewBadRequestFromMessage(message string) *ApiError {
	return NewBadRequestError(BAD_REQUEST_CODE, message)
}

func NewForbiddenError(code string, message string) *ApiError {
	return NewCustomError(FORBIDDEN, code, message)
}

func NewValidationError(message string, errors ...ErrorDetails) *ApiError {
	return NewCustomError(BAD_REQUEST, INVALID_INPUT_CODE, message, errors...)
}

func GenerateErrorFromStatus(status int) *ApiError {
	switch status {
	case UNAUTHORIZED:
		return NewUnAuthorizedError("Unauthorized")
	case NOT_FOUND:
		return NewNotFoundError("Not Found")
	case BAD_REQUEST:
		return NewBadRequestError("Bad Request", BAD_REQUEST_CODE)
	case FORBIDDEN:
		return NewForbiddenError("Forbidden", FORBIDDEN_CODE)
	default:
		return NewInternalServerError()
	}
}
