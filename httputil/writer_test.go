package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/stretchr/testify/require"
)

var noopListener OnErrorListener = func(err error) {}

type testPayload struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func newWriter(w *httptest.ResponseRecorder) JsonResponseWriter {
	return NewJsonResponseWriter(w).WithErrorListener(noopListener)
}

func decodeBody[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var out T
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&out))
	return out
}

// --- WriteData ---

func TestWriteData_Success(t *testing.T) {
	rec := httptest.NewRecorder()
	data := testPayload{Name: "hello", Value: 42}
	newWriter(rec).WriteData(data)

	require.Equal(t, http.StatusOK, rec.Code)
	got := decodeBody[testPayload](t, rec)
	require.Equal(t, data.Name, got.Name)
	require.Equal(t, data.Value, got.Value)
}

func TestWriteData_CustomSuccessStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	newWriter(rec).WithSuccessStatus(http.StatusCreated).WriteData(testPayload{Name: "created"})

	require.Equal(t, http.StatusCreated, rec.Code)
}

// --- WriteJsonResponse ---

func TestWriteJsonResponse_WithData(t *testing.T) {
	rec := httptest.NewRecorder()
	data := testPayload{Name: "resp", Value: 7}
	newWriter(rec).WriteJsonResponse(data, nil)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	got := decodeBody[testPayload](t, rec)
	require.Equal(t, data.Name, got.Name)
	require.Equal(t, data.Value, got.Value)
}

func TestWriteJsonResponse_WithError(t *testing.T) {
	rec := httptest.NewRecorder()
	newWriter(rec).WriteJsonResponse(nil, &ResultError{Message: "boom", Code: "UNKNOWN_ERROR"})

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

// --- WriteSuccessMessage ---

func TestWriteSuccessMessage(t *testing.T) {
	rec := httptest.NewRecorder()
	newWriter(rec).WriteSuccessMessage("ok", nil)

	got := decodeBody[map[string]string](t, rec)
	require.Equal(t, "ok", got["message"])
}

// --- WriteError: random (non-ResultError) error ---

func TestWriteError_RandomError(t *testing.T) {
	rec := httptest.NewRecorder()
	newWriter(rec).WriteError(&ResultError{Message: "something went wrong", Code: "INTERNAL_ERROR"})

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	got := decodeBody[ResultError](t, rec)
	require.Equal(t, CodeInternal, got.Code)
}

// --- WriteError: table-driven across all ResultError codes ---

func TestWriteError_CustomResultError(t *testing.T) {
	cases := []struct {
		name       string
		err        *ResultError
		wantStatus int
		wantCode   string
	}{
		{"BadRequest", BadRequestError("msg"), http.StatusBadRequest, CodeBadRequest},
		{"Unauthorized", UnauthorizedError("msg"), http.StatusUnauthorized, CodeUnauthorized},
		{"Forbidden", ForbiddenError("msg"), http.StatusForbidden, CodeForbidden},
		{"NotFound", NotFoundError("msg"), http.StatusNotFound, CodeNotFound},
		{"Internal", InternalError("msg"), http.StatusInternalServerError, CodeInternal},
		{"Unknown", UnknownError("msg"), http.StatusInternalServerError, CodeUnknown},
		{"InvalidInput", InvalidInputError("msg"), http.StatusUnprocessableEntity, CodeInvalidInput},
		{"Validation", NewValidationError(), http.StatusUnprocessableEntity, CodeValidation},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			newWriter(rec).WriteError(tc.err)

			require.Equal(t, tc.wantStatus, rec.Code)
			got := decodeBody[ResultError](t, rec)
			require.Equal(t, tc.wantCode, got.Code)
		})
	}
}

// --- WriteError: custom error status override ---

func TestWriteError_CustomErrorStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	newWriter(rec).WithErrorStatus(http.StatusServiceUnavailable).WriteError(NotFoundError("x"))

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

// --- WithErrorListener ---

func TestOnErrorListener(t *testing.T) {
	rec := httptest.NewRecorder()
	var captured error
	listener := OnErrorListener(func(err error) { captured = err })

	sentErr := NotFoundError("trigger listener")
	NewJsonResponseWriter(rec).WithErrorListener(listener).WriteError(sentErr)

	require.NotNil(t, captured)
	require.Equal(t, sentErr.Error(), captured.Error())
}

// --- WriteError: response body fields ---

func TestWriteError_ResponseBodyContainsMessage(t *testing.T) {
	rec := httptest.NewRecorder()
	newWriter(rec).WriteError(NotFoundError("item not found"))

	got := decodeBody[ResultError](t, rec)
	require.Equal(t, "item not found", got.Message)
	require.Equal(t, CodeNotFound, got.Code)
}
