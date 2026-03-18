package grpcutil

import (
	"context"
	"io"
	"log/slog"
	"testing"

	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/errorspb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

var fakeInfo = &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

// --- RecoverUnary: no panic ---

func TestRecoverUnary_NoPanic(t *testing.T) {
	interceptor := RecoverUnary()
	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), nil, fakeInfo, handler)
	require.NoError(t, err)
	require.Equal(t, "ok", resp)
}

// --- RecoverUnary: panic with error ---

func TestRecoverUnary_PanicWithError(t *testing.T) {
	interceptor := RecoverUnary(RecoveryOptions{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	handler := func(ctx context.Context, req any) (any, error) {
		panic(NotFoundError("x"))
	}

	_, err := interceptor(context.Background(), nil, fakeInfo, handler)
	pbErr, ok := err.(*errorspb.ResultErrorPb)
	require.True(t, ok)
	require.Equal(t, CodeNotFound, pbErr.Code)
}

// --- RecoverUnary: panic with non-error string ---

func TestRecoverUnary_PanicWithString(t *testing.T) {
	interceptor := RecoverUnary(RecoveryOptions{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	handler := func(ctx context.Context, req any) (any, error) {
		panic("something went wrong")
	}

	_, err := interceptor(context.Background(), nil, fakeInfo, handler)
	pbErr, ok := err.(*errorspb.ResultErrorPb)
	require.True(t, ok)
	require.Equal(t, CodeInternal, pbErr.Code)
}

// --- RecoverUnary: custom logger option ---

func TestRecoverUnary_CustomLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	interceptor := RecoverUnary(RecoveryOptions{Logger: logger})
	handler := func(ctx context.Context, req any) (any, error) {
		panic("custom logger test")
	}

	_, err := interceptor(context.Background(), nil, fakeInfo, handler)
	require.Error(t, err)
}

// --- RecoverUnary: default logger (no opts) ---

func TestRecoverUnary_DefaultLogger(t *testing.T) {
	interceptor := RecoverUnary()
	handler := func(ctx context.Context, req any) (any, error) {
		panic("default logger test")
	}

	_, err := interceptor(context.Background(), nil, fakeInfo, handler)
	require.Error(t, err)
}
