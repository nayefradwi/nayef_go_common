package connectutil

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

const procedure = "/test.v1.TestService/Call"

// Sends handlerErr over a real connect round trip and returns what the client sees.
func callWith(t *testing.T, handlerErr error, serverOpts ErrorOptions, clientOpts ...connect.ClientOption) error {
	t.Helper()

	handler := connect.NewUnaryHandler(
		procedure,
		func(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
			return nil, handlerErr
		},
		connect.WithInterceptors(ServerErrors(serverOpts)),
	)

	mux := http.NewServeMux()
	mux.Handle(procedure, handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := connect.NewClient[emptypb.Empty, emptypb.Empty](server.Client(), server.URL+procedure, clientOpts...)
	_, err := client.CallUnary(context.Background(), connect.NewRequest(&emptypb.Empty{}))
	return err
}

func callResult(t *testing.T, handlerErr error) *ResultError {
	t.Helper()

	err := callWith(t, handlerErr, ErrorOptions{ErrorListener: func(error) {}}, connect.WithInterceptors(ClientErrors()))

	var resultErr *ResultError
	require.True(t, errors.As(err, &resultErr), "got %T: %v", err, err)
	return resultErr
}

func TestServerErrors_MasksUnknownError(t *testing.T) {
	var seen error
	opts := ErrorOptions{ErrorListener: func(err error) { seen = err }}
	leak := errors.New("pq: password=secret")

	err := callWith(t, leak, opts, connect.WithInterceptors(ClientErrors()))

	var resultErr *ResultError
	require.True(t, errors.As(err, &resultErr))
	require.Equal(t, CodeInternal, resultErr.Code)
	require.Equal(t, "internal server error", resultErr.Message)
	require.NotContains(t, err.Error(), "secret")
	require.Equal(t, leak, seen)
}

func TestServerErrors_WrappedResultError(t *testing.T) {
	var notFound error = NotFoundError("x")
	out := callResult(t, fmt.Errorf("wrap: %w", notFound))

	require.Equal(t, CodeNotFound, out.Code)
	require.Equal(t, http.StatusNotFound, out.Status)
	require.Equal(t, "x", out.Message)
}

func TestServerErrors_DetailsSurviveWire(t *testing.T) {
	in := NewValidationError(
		Field("email", "REQUIRED", "email is required"),
		Field("email", "FORMAT", "email is malformed"),
	)

	out := callResult(t, in)

	require.Equal(t, CodeValidation, out.Code)
	require.Equal(t, http.StatusUnprocessableEntity, out.Status)
	require.Equal(t, in.Errors, out.Errors)
}

func TestServerErrors_ConnectErrorPassesThrough(t *testing.T) {
	out := callResult(t, connect.NewError(connect.CodeUnavailable, errors.New("try later")))

	require.Equal(t, "UNAVAILABLE", out.Code)
	require.Equal(t, "try later", out.Message)
}

func TestClientWithoutInterceptor_GetsConnectError(t *testing.T) {
	err := callWith(t, NotFoundError("x"), ErrorOptions{ErrorListener: func(error) {}})

	var connectErr *connect.Error
	require.True(t, errors.As(err, &connectErr))
	require.Equal(t, connect.CodeNotFound, connectErr.Code())

	var resultErr *ResultError
	require.False(t, errors.As(err, &resultErr))
}
