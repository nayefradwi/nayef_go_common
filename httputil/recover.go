package httputil

import (
	"fmt"
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
	if recovered == nil {
		return
	}

	jw := NewJsonResponseWriter(w)
	err, ok := recovered.(error)
	if !ok || err == nil {
		err = fmt.Errorf("%v", recovered)
	}

	jw.WriteError(err)
}
