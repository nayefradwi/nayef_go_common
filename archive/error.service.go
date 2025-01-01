package common

import (
	"net/http"

	"go.uber.org/zap"
)

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
