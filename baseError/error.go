package baseError

type BaseError struct {
	Message string                 `json:"message"`
	Status  int                    `json:"status"`
	Fields  []ValidationFieldError `json:"fields,omitempty"`
}

type ValidationFieldError struct {
	Message string `json:"message"`
	Field   string `json:"field"`
}

func (e BaseError) Error() string {
	return e.Message
}
