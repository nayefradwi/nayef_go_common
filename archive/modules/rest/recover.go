package rest

import (
	"net/http"

	"go.uber.org/zap"
)

func Recover(f http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverError(w)
		f.ServeHTTP(w, r)
	})

	return handler
}

func recoverError(w http.ResponseWriter) {
	recovered := recover()
	if err, ok := recovered.(error); ok {
		zap.L().Error("internal server error", zap.Any("error", err), zap.Stack("stack trace"))
		jw := NewJsonResponseWriter(w)
		jw.WriteError(err)
	}
}
