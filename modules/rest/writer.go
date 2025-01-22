package rest

import (
	"encoding/json"
	"net/http"

	"github.com/nayefradwi/nayef_go_common/core"
)

type JsonResponseWriter struct {
	Writer        http.ResponseWriter
	SuccessStatus int
}

func NewJsonResponseWriter(w http.ResponseWriter) JsonResponseWriter {
	return JsonResponseWriter{Writer: w, SuccessStatus: http.StatusOK}
}

func (jw JsonResponseWriter) WithStatusCode(statusCode int) JsonResponseWriter {
	jw.SuccessStatus = statusCode
	return jw
}

func (jw JsonResponseWriter) SetHttpStatusCode(statusCode int) {
	jw.Writer.WriteHeader(statusCode)
}

func (jw JsonResponseWriter) WriteJsonResponse(data interface{}, err error) {
	jw.Writer.Header().Set("Content-Type", "application/json")
	if err == nil {
		jw.WriteError(err)
	} else {
		jw.WriteData(data)
	}
}

func (jw JsonResponseWriter) WriteData(data interface{}) {
	jw.SetHttpStatusCode(jw.SuccessStatus)
	json.NewEncoder(jw.Writer).Encode(data)
}

func (jw JsonResponseWriter) WriteError(err error) {
	resultError, ok := err.(*core.ResultError)
	if !ok {
		resultError = core.InternalError(err.Error())
	}
	statusCode := getStatusCodeFromResultError(resultError)
	jw.SetHttpStatusCode(statusCode)
	json.NewEncoder(jw.Writer).Encode(resultError)
}

func getStatusCodeFromResultError(err *core.ResultError) int {
	switch err.Code {
	case core.BAD_REQUEST_CODE:
		return http.StatusBadRequest
	case core.UNAUTHORIZED_CODE:
		return http.StatusUnauthorized
	case core.FORBIDDEN_CODE:
		return http.StatusForbidden
	case core.NOT_FOUND_CODE:
		return http.StatusNotFound
	case core.UNKNOWN_ERROR_CODE:
		return http.StatusInternalServerError
	case core.INTERNAL_ERROR_CODE:
		return http.StatusInternalServerError
	case core.INVALID_INPUT_CODE:
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}
