package response

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/nayefradwi/nayef_go_common/baseError"
)

type successCallback[T any] func(T)

func ParseBody[T any](w http.ResponseWriter, body io.ReadCloser, onSuccess successCallback[T]) {
	var data T
	err := json.NewDecoder(body).Decode(&data)
	if err != nil {
		log.Printf("failed to parse body: %s", err.Error())
		WriteErrorResponse(w, baseError.NewInternalServerError().(*baseError.BaseError))
		return
	}
	onSuccess(data)
}
