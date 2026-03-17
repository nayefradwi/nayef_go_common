package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nayefradwi/nayef_go_common/result"
)

var (
	GlobalJsonWriterOnErrorListener result.OnErrorListener = func(err error) {}
)

func SuccessJsonResponseMessage(message string) map[string]string {
	return map[string]string{"message": message}
}

type JsonResponseWriter struct {
	Writer        http.ResponseWriter
	SuccessStatus int
	ErrorStatus   int
	ErrorListener result.OnErrorListener
}

func NewJsonResponseWriter(w http.ResponseWriter) JsonResponseWriter {
	return JsonResponseWriter{Writer: w, SuccessStatus: http.StatusOK, ErrorListener: GlobalJsonWriterOnErrorListener}
}

func (jw JsonResponseWriter) WithSuccessStatus(statusCode int) JsonResponseWriter {
	jw.SuccessStatus = statusCode
	return jw
}

func (jw JsonResponseWriter) WithErrorStatus(status int) JsonResponseWriter {
	jw.ErrorStatus = status
	return jw
}

func (jw JsonResponseWriter) SetHttpStatusCode(statusCode int) {
	jw.Writer.WriteHeader(statusCode)
}

func (jw JsonResponseWriter) WriteJsonResponse(data interface{}, err error) {
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

func (jw JsonResponseWriter) WriteData(data interface{}) {
	jw.SetHttpStatusCode(jw.SuccessStatus)
	json.NewEncoder(jw.Writer).Encode(data)
}

func (jw JsonResponseWriter) WriteError(err error) {
	jw.ErrorListener(err)
	var resultError *result.ResultError
	if !errors.As(err, &resultError) {
		resultError = result.InternalError(err.Error())
	}

	statusCode := jw.ErrorStatus
	if statusCode < 400 || statusCode > 505 {
		statusCode = getStatusCodeFromResultError(resultError)
	}

	jw.SetHttpStatusCode(statusCode)
	json.NewEncoder(jw.Writer).Encode(resultError)
}

func getStatusCodeFromResultError(err *result.ResultError) int {
	switch err.Code {
	case result.BAD_REQUEST_CODE:
		return http.StatusBadRequest
	case result.UNAUTHORIZED_CODE:
		return http.StatusUnauthorized
	case result.FORBIDDEN_CODE:
		return http.StatusForbidden
	case result.NOT_FOUND_CODE:
		return http.StatusNotFound
	case result.UNKNOWN_ERROR_CODE:
		return http.StatusInternalServerError
	case result.INTERNAL_ERROR_CODE:
		return http.StatusInternalServerError
	case result.INVALID_INPUT_CODE:
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}
