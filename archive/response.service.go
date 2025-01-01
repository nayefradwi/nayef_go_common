package common

import (
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type EmptyResult struct {
	Writer  http.ResponseWriter
	Message string
	Error   error
}

type Result[T any] struct {
	Writer http.ResponseWriter
	Data   T
	Error  error
}

type SuccessCallback[T any] func(T)

func ParseBody[T any](w http.ResponseWriter, body io.ReadCloser, onSuccess SuccessCallback[T]) {
	var data T
	err := json.NewDecoder(body).Decode(&data)
	if err != nil {
		GetLogger().Error("failed to parse body", zap.Error(err))
		WriteResponseFromError(w, NewInternalServerError())
		return
	}
	GetLogger().Info("successfully parsed body", zap.Any("data", data))
	onSuccess(data)
}

/*
example to use this is:

	ParseBody[TYPE](w, r.Body, func(data TYPE){
		...DO SOMETHING WITH DATA
		WriteResponse[TYPE](Result[TYPE]{Writer: w, Data: data, Error: err})
	})

in this case if the error is nil it will write the body as json
if the error is an api error then the error will be written as json
this creates a consistent way of handling 3 cases:
  - api returns data
  - api returns success / created without data
  - api returns error

so the pattern is always:
  - parse body if needed
  - call service which calls repo
  - write response
*/
func (result Result[T]) WriteResponse() {
	if result.Error != nil {
		WriteResponseFromError(result.Writer, result.Error)
		return
	}
	GetLogger().Info("successfully wrote response", zap.Any("data", result.Data))
	json.NewEncoder(result.Writer).Encode(result.Data)
}

func (result EmptyResult) WriteCreatedResponse() {
	if result.Error != nil {
		WriteResponseFromError(result.Writer, result.Error)
		return
	}
	writeResponse(result, 201)

}
func (result EmptyResult) WriteEmptyResponse() {
	if result.Error != nil {
		WriteResponseFromError(result.Writer, result.Error)
		return
	}
	writeResponse(result, 200)
}

func writeResponse(result EmptyResult, status int) {
	body := make(map[string]interface{})
	body["message"] = result.Message
	body["status"] = status
	GetLogger().Info("successfully wrote response", zap.Any("data", body))
	json.NewEncoder(result.Writer).Encode(body)
}

func WriteResponseFromError(w http.ResponseWriter, e error) {
	defer handleInternalError(w)
	GetLogger().Error("responding with error", zap.Error(e))
	if apiError, ok := e.(*ApiError); ok {
		w.WriteHeader(apiError.Status)
		w.Write(apiError.GenerateResponse())
	}
}

func handleInternalError(w http.ResponseWriter) {
	err := recover()
	if err != nil {
		GetLogger().Error("internal server error", zap.Any("error", err), zap.Stack("stack trace"))
		err := NewInternalServerError()
		response := err.GenerateResponse()
		w.WriteHeader(err.Status)
		w.Write(response)
	}
}

func JsonResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
