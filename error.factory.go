package common

func newError(message string, status int, code string) *ApiError {
	return &ApiError{
		Message: message,
		Status:  status,
		Code:    code,
	}
}

func NewCustomError(message string, status int, code string, errors ...ErrorDetails) *ApiError {
	return &ApiError{
		Message: message,
		Status:  status,
		Code:    code,
		Errors:  errors,
	}
}

func NewUnAuthorizedError(message string) *ApiError {
	return newError(message, UNAUTHORIZED, UNAUTHORIZED_CODE)
}

func NewInternalServerError() *ApiError {
	return newError("Internal server error", INTERNAL_SERVER_ERROR, INTERNAL_ERROR_CODE)
}

func NewNotFoundError(message string) *ApiError {
	return newError(message, NOT_FOUND, NOT_FOUND_CODE)
}

func NewBadRequestError(message string, code string) *ApiError {
	return NewCustomError(message, BAD_REQUEST, code)
}

func NewBadRequestFromMessage(message string) *ApiError {
	return NewBadRequestError(message, BAD_REQUEST_CODE)
}

func NewForbiddenError(message string, code string) *ApiError {
	return NewCustomError(message, FORBIDDEN, code)
}

func NewValidationError(message string, errors ...ErrorDetails) *ApiError {
	return NewCustomError(message, BAD_REQUEST, INVALID_INPUT_CODE, errors...)
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
