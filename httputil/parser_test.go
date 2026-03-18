package httputil

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseJsonBody_Success(t *testing.T) {
	data := testPayload{Name: "parse me", Value: 99}
	b, _ := json.Marshal(data)
	body := io.NopCloser(strings.NewReader(string(b)))

	rec := httptest.NewRecorder()
	called := false

	ParseJsonBody(rec, body, func(jw JsonResponseWriter, parsed testPayload) {
		called = true
		require.Equal(t, data.Name, parsed.Name)
		require.Equal(t, data.Value, parsed.Value)
	})

	require.True(t, called, "expected onSuccess to be called")
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestParseJsonBody_InvalidJson(t *testing.T) {
	body := io.NopCloser(strings.NewReader("not-json"))

	rec := httptest.NewRecorder()
	called := false
	// Override global listener to suppress slog noise during the error path.
	orig := GlobalJsonWriterOnErrorListener
	GlobalJsonWriterOnErrorListener = noopListener
	defer func() { GlobalJsonWriterOnErrorListener = orig }()

	ParseJsonBody(rec, body, func(jw JsonResponseWriter, parsed testPayload) {
		called = true
	})

	require.False(t, called, "expected onSuccess NOT to be called on invalid JSON")
	require.NotEqual(t, http.StatusOK, rec.Code)

	var errBody map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&errBody))
}
