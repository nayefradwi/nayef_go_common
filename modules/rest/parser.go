package rest

import (
	"encoding/json"
	"io"
	"net/http"
)

type SuccessParsingCallback[T any] func(jw JsonResponseWriter, data T)

func ParseJsonBody[T any](w http.ResponseWriter, body io.ReadCloser, onSuccess SuccessParsingCallback[T]) {
	var data T
	jw := NewJsonResponseWriter(w)
	err := json.NewDecoder(body).Decode(&data)
	if err != nil {
		jw.writeError(err)
		return
	}

	onSuccess(jw, data)
}
