package common

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func (e ApiError) GenerateResponse() []byte {
	errorResponse, err := json.Marshal(e)
	if err != nil {
		internalServerError, _ := json.Marshal(NewInternalServerError())
		return internalServerError
	}
	return errorResponse
}

func Recover(f http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				GetLogger().Error(
					"internal error in commons package",
					zap.Any("error", err),
					zap.Stack("stack trace"),
				)
				err := NewInternalServerError()
				result := Result[interface{}]{Error: err, Writer: w}
				result.WriteResponse()
			}
		}()
		f.ServeHTTP(w, r)
	})
	return handler
}
