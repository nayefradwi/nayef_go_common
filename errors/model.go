package errors

type OnErrorListener func(err error)

type ErrorDetails struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Field   string `json:"field"`
}

type ResultError struct {
	Message string                    `json:"message"`
	Code    string                    `json:"code"`
	Errors  map[string][]ErrorDetails `json:"errors,omitempty,omitzero"`
}

func (e ResultError) Error() string {
	return e.Message
}

func (e ResultError) WithCode(code string) ResultError {
	e.Code = code
	return e
}

func (e ResultError) WithErrors(details ...ErrorDetails) ResultError {
	errs := make(map[string][]ErrorDetails)
	for _, d := range details {
		errs[d.Field] = append(errs[d.Field], d)
	}
	e.Errors = errs
	return e
}

func NewResultError(message string, code string, details ...ErrorDetails) *ResultError {
	errs := make(map[string][]ErrorDetails)
	for _, d := range details {
		errs[d.Field] = append(errs[d.Field], d)
	}
	return &ResultError{
		Message: message,
		Code:    code,
		Errors:  errs,
	}
}
