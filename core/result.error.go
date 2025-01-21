package core

const (
	NOT_FOUND_CODE        = "NOT_FOUND"
	BAD_REQUEST_CODE      = "BAD_REQUEST"
	UNAUTHORIZED_CODE     = "UNAUTHORIZED"
	FORBIDDEN_CODE        = "FORBIDDEN"
	INTERNAL_ERROR_CODE   = "INTERNAL_ERROR"
	INVALID_INPUT_CODE    = "INVALID_INPUT"
	UNKNOWN_ERROR_CODE    = "UNKNOWN_ERROR"
	VALIDATION_ERROR_CODE = "VALIDATION_ERROR"
)

type ResultError struct {
	Message string         `json:"message"`
	Code    string         `json:"code"`
	Errors  []ErrorDetails `json:"errors,omitempty"`
}

type ErrorDetails struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
	Field   string `json:"field,omitempty"`
}

func (e ResultError) Error() string {
	return e.Message
}

func NewResultError(message string, code string, details ...ErrorDetails) *ResultError {
	return &ResultError{
		Message: message,
		Code:    code,
		Errors:  details,
	}
}

func (e ResultError) WithCode(code string) ResultError {
	e.Code = code
	return e
}

func (e ResultError) WithErrorDetails(details []ErrorDetails) ResultError {
	e.Errors = details
	return e
}
