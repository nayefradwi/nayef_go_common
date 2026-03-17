package httputil

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	. "github.com/nayefradwi/nayef_go_common/errors"
)

var (
	GlobalJsonWriterOnErrorListener OnErrorListener = func(err error) {
		slog.Error("GlobalJsonWriterOnErrorListener", "error", err.Error())
	}
)

func SuccessJsonResponseMessage(message string) map[string]string {
	return map[string]string{"message": message}
}

type JsonResponseWriter struct {
	Writer        http.ResponseWriter
	SuccessStatus int
	ErrorStatus   int
	ErrorListener OnErrorListener
}

func NewJsonResponseWriter(w http.ResponseWriter) JsonResponseWriter {
	return JsonResponseWriter{Writer: w, SuccessStatus: http.StatusOK, ErrorListener: GlobalJsonWriterOnErrorListener}
}

func (jw JsonResponseWriter) WithSuccessStatus(statusCode int) JsonResponseWriter {
	jw.SuccessStatus = statusCode
	return jw
}

func (jw JsonResponseWriter) WithErrorListener(listener OnErrorListener) JsonResponseWriter {
	jw.ErrorListener = listener
	return jw
}

func (jw JsonResponseWriter) WithErrorStatus(status int) JsonResponseWriter {
	jw.ErrorStatus = status
	return jw
}

func (jw JsonResponseWriter) SetHttpStatusCode(statusCode int) {
	jw.Writer.WriteHeader(statusCode)
}

func (jw JsonResponseWriter) WriteJsonResponse(data any, err error) {
	jw.Writer.Header().Set("Content-Type", "application/json")
	if err != nil {
		jw.WriteError(err)
	} else {
		jw.WriteData(data)
	}
}

func (jw JsonResponseWriter) WriteSuccessMessage(data string, err error) {
	messageResponse := SuccessJsonResponseMessage(data)
	jw.WriteJsonResponse(messageResponse, err)
}

func (jw JsonResponseWriter) WriteData(data any) {
	jw.SetHttpStatusCode(jw.SuccessStatus)
	json.NewEncoder(jw.Writer).Encode(data)
}

func (jw JsonResponseWriter) WriteError(err error) {
	jw.ErrorListener(err)
	var resultError *ResultError
	if !errors.As(err, &resultError) {
		resultError = InternalError(err.Error())
	}

	statusCode := jw.ErrorStatus
	if statusCode < 400 || statusCode > 505 {
		statusCode = getStatusCodeFromResultError(resultError)
	}

	jw.SetHttpStatusCode(statusCode)
	json.NewEncoder(jw.Writer).Encode(resultError)
}

func getStatusCodeFromResultError(err *ResultError) int {
	switch err.Code {
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeUnknown, CodeInternal:
		return http.StatusInternalServerError
	case CodeInvalidInput, CodeValidation:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusBadRequest
	}
}
