package grpcutil

import (
	"fmt"
	"testing"

	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/errorspb"
	"github.com/stretchr/testify/require"
)

var noopListener OnErrorListener = func(err error) {}

func newWriter[T any]() GrpcResponseWriter[T] {
	return NewGrpcResponseWriter[T]().WithErrorListener(noopListener)
}

// --- WriteData ---

func TestWriteData_ReturnsDataAndNilError(t *testing.T) {
	data := "hello"
	got, err := newWriter[string]().WriteData(data)
	require.NoError(t, err)
	require.Equal(t, data, got)
}

// --- WriteError: plain error wrapped as Internal ---

func TestWriteError_PlainError_WrappedAsInternal(t *testing.T) {
	err := newWriter[any]().WriteError(fmt.Errorf("boom"))
	pbErr, ok := err.(*errorspb.ResultErrorPb)
	require.True(t, ok)
	require.Equal(t, CodeInternal, pbErr.Code)
}

// --- WriteError: table-driven over all ResultError codes ---

func TestWriteError_ResultError(t *testing.T) {
	cases := []struct {
		name     string
		err      *ResultError
		wantCode string
		wantMsg  string
	}{
		{"BadRequest", BadRequestError("msg"), CodeBadRequest, "msg"},
		{"Unauthorized", UnauthorizedError("msg"), CodeUnauthorized, "msg"},
		{"Forbidden", ForbiddenError("msg"), CodeForbidden, "msg"},
		{"NotFound", NotFoundError("msg"), CodeNotFound, "msg"},
		{"Internal", InternalError("msg"), CodeInternal, "msg"},
		{"Unknown", UnknownError("msg"), CodeUnknown, "msg"},
		{"InvalidInput", InvalidInputError("msg"), CodeInvalidInput, "msg"},
		{"Validation", NewValidationError(), CodeValidation, "Invalid"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := newWriter[any]().WriteError(tc.err)
			pbErr, ok := err.(*errorspb.ResultErrorPb)
			require.True(t, ok)
			require.Equal(t, tc.wantCode, pbErr.Code)
			require.Equal(t, tc.wantMsg, pbErr.Message)
		})
	}
}

// --- WriteError: listener is called ---

func TestWriteError_ListenerCalled(t *testing.T) {
	var captured error
	listener := OnErrorListener(func(err error) { captured = err })
	sentErr := NotFoundError("trigger")

	NewGrpcResponseWriter[any]().WithErrorListener(listener).WriteError(sentErr)

	require.NotNil(t, captured)
	require.Equal(t, sentErr.Error(), captured.Error())
}

// --- WriteError: global listener fallback ---

func TestWriteError_GlobalListener(t *testing.T) {
	orig := GlobalWriterOnErrorListener
	defer func() { GlobalWriterOnErrorListener = orig }()

	var captured error
	GlobalWriterOnErrorListener = func(err error) { captured = err }

	sentErr := InternalError("global")
	NewGrpcResponseWriter[any]().WriteError(sentErr)

	require.NotNil(t, captured)
	require.Equal(t, sentErr.Error(), captured.Error())
}

// --- WriteResponse ---

func TestWriteResponse_WithError(t *testing.T) {
	_, err := newWriter[string]().WriteResponse("", NotFoundError("not found"))
	pbErr, ok := err.(*errorspb.ResultErrorPb)
	require.True(t, ok)
	require.Equal(t, CodeNotFound, pbErr.Code)
}

func TestWriteResponse_NoError(t *testing.T) {
	data := "success"
	got, err := newWriter[string]().WriteResponse(data, nil)
	require.NoError(t, err)
	require.Equal(t, data, got)
}

// --- WithErrorListener ---

func TestWithErrorListener_OverridesGlobal(t *testing.T) {
	orig := GlobalWriterOnErrorListener
	defer func() { GlobalWriterOnErrorListener = orig }()

	globalCalled := false
	GlobalWriterOnErrorListener = func(err error) { globalCalled = true }

	var localCaptured error
	localListener := OnErrorListener(func(err error) { localCaptured = err })

	sentErr := BadRequestError("override")
	NewGrpcResponseWriter[any]().WithErrorListener(localListener).WriteError(sentErr)

	require.NotNil(t, localCaptured)
	require.False(t, globalCalled)
}
