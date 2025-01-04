package rest

import (
	"net/http"
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
		jw := NewJsonResponseWriter(w)
		jw.writeError(err)
	}
}
