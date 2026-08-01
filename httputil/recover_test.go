package httputil

import (
	goerrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/stretchr/testify/require"
)

func serveRecovered(h http.HandlerFunc) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	Recover(h).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	return rec
}

// recover() returns nil on the normal path too, so the handler's own response
// must not have an error body appended to it.
func TestRecover_NoPanic_LeavesResponseUntouched(t *testing.T) {
	rec := serveRecovered(func(w http.ResponseWriter, _ *http.Request) {
		NewJsonResponseWriter(w).
			WithErrorListener(noopListener).
			WriteJsonResponse(testPayload{Name: "hello", Value: 42}, nil)
	})

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"name":"hello","value":42}`, rec.Body.String())
}

func TestRecover_NoPanic_HandlerWritesNothing(t *testing.T) {
	rec := serveRecovered(func(http.ResponseWriter, *http.Request) {})

	require.Equal(t, http.StatusOK, rec.Code)
	require.Empty(t, rec.Body.String())
}

func TestRecover_PanicWithError(t *testing.T) {
	rec := serveRecovered(func(http.ResponseWriter, *http.Request) {
		panic(goerrors.New("boom"))
	})

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, CodeInternal, decodeBody[ResultError](t, rec).Code)
}

func TestRecover_PanicWithString(t *testing.T) {
	rec := serveRecovered(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, CodeInternal, decodeBody[ResultError](t, rec).Code)
}

func TestRecover_PanicWithNonErrorValue(t *testing.T) {
	rec := serveRecovered(func(http.ResponseWriter, *http.Request) {
		panic(42)
	})

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

// A panicked ResultError keeps its own status and message rather than being
// flattened into a generic internal error.
func TestRecover_PanicWithResultError(t *testing.T) {
	rec := serveRecovered(func(http.ResponseWriter, *http.Request) {
		panic(NotFoundError("user not found"))
	})

	require.Equal(t, http.StatusNotFound, rec.Code)
	body := decodeBody[ResultError](t, rec)
	require.Equal(t, CodeNotFound, body.Code)
	require.Equal(t, "user not found", body.Message)
}

func TestRecover_PanicSetsJsonContentType(t *testing.T) {
	rec := serveRecovered(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}
