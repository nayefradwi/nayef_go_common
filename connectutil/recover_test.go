package connectutil

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Runs a handler that calls panicker over a real connect round trip; returns the client error and the log output.
func callPanic(t *testing.T, panicker func()) (*ResultError, string) {
	t.Helper()

	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	handler := connect.NewUnaryHandler(
		procedure,
		func(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
			panicker()
			return connect.NewResponse(&emptypb.Empty{}), nil
		},
		connect.WithRecover(Recover(RecoveryOptions{Logger: logger})),
		connect.WithInterceptors(ServerErrors(ErrorOptions{ErrorListener: func(error) {}})),
	)

	mux := http.NewServeMux()
	mux.Handle(procedure, handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := connect.NewClient[emptypb.Empty, emptypb.Empty](server.Client(), server.URL+procedure,
		connect.WithInterceptors(ClientErrors()))
	_, err := client.CallUnary(context.Background(), connect.NewRequest(&emptypb.Empty{}))

	resultErr, ok := errors.AsType[*ResultError](err)
	require.True(t, ok, "got %T: %v", err, err)
	return resultErr, logs.String()
}

func requireMasked(t *testing.T, out *ResultError) {
	t.Helper()
	require.Equal(t, CodeInternal, out.Code)
	require.Equal(t, "internal server error", out.Message)
	require.Equal(t, http.StatusInternalServerError, out.Status)
}

func TestRecover_MasksPanicAndLogsCause(t *testing.T) {
	out, logs := callPanic(t, func() { panic("db password=secret") })

	requireMasked(t, out)
	require.NotContains(t, out.Error(), "secret")
	require.Contains(t, logs, "secret")
	require.Contains(t, logs, procedure)
	require.Contains(t, logs, "stack=")
}

func TestRecover_ResultErrorPassesThrough(t *testing.T) {
	out, _ := callPanic(t, func() { panic(NotFoundError("x")) })

	require.Equal(t, CodeNotFound, out.Code)
	require.Equal(t, http.StatusNotFound, out.Status)
	require.Equal(t, "x", out.Message)
}

func TestRecover_RuntimePanicMasked(t *testing.T) {
	out, logs := callPanic(t, func() {
		var nilMap map[string]int
		nilMap["a"] = 1
	})

	requireMasked(t, out)
	require.Contains(t, logs, "nil map")
}

func TestRecover_NilPanicMasked(t *testing.T) {
	out, _ := callPanic(t, func() { panic(nil) })

	requireMasked(t, out)
}
