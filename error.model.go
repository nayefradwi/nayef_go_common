package common

// only for common errors the codes are like their status codes
// for custom errors the code will be based on the operation
const (
	NOT_FOUND             int = 404
	BAD_REQUEST           int = 400
	UNAUTHORIZED          int = 401
	FORBIDDEN             int = 403
	INTERNAL_SERVER_ERROR int = 500
)

const (
	NOT_FOUND_CODE      = "NOT_FOUND"
	BAD_REQUEST_CODE    = "BAD_REQUEST"
	UNAUTHORIZED_CODE   = "UNAUTHORIZED"
	FORBIDDEN_CODE      = "FORBIDDEN"
	INTERNAL_ERROR_CODE = "INTERNAL_ERROR"
	INVALID_INPUT_CODE  = "INVALID_INPUT"
	UNKNOWN_ERROR_CODE  = "UNKNOWN_ERROR"
)

type ApiError struct {
	Message string         `json:"message"`
	Status  int            `json:"status"`
	Code    string         `json:"code,omitempty"`
	Errors  []ErrorDetails `json:"errors,omitempty"`
}

type ErrorDetails struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
	Field   string `json:"field,omitempty"`
}

func (e *ApiError) Error() string {
	return e.Message
}
