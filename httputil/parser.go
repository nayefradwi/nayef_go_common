package httputil

import (
	"encoding/json/v2"
	"io"
	"net/http"

	. "github.com/nayefradwi/nayef_go_common/errors"
)

type SuccessParsingCallback[T any] func(jw JsonResponseWriter, data T)

func ParseJsonBody[T any](w http.ResponseWriter, body io.ReadCloser, onSuccess SuccessParsingCallback[T]) {
	var data T
	jw := NewJsonResponseWriter(w)
	if err := json.UnmarshalRead(body, &data); err != nil {
		// joined so the error listener still logs the decode cause; the client only sees the bad request
		jw.WriteError(Join(BadRequestError("invalid json body"), err))
		return
	}

	onSuccess(jw, data)
}
