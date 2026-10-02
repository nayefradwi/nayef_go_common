package connectutil

import (
	"errors"
	"net/http"
	"testing"

	"connectrpc.com/connect"
	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/stretchr/testify/require"
)

func TestToConnectError_Nil(t *testing.T) {
	require.Nil(t, ToConnectError(nil))
	require.Nil(t, ToResultError(nil))
}

// NewValidationError sets no status; it must still become InvalidArgument and come back as 422.
func TestValidationError_RoundTrip(t *testing.T) {
	in := NewValidationError(
		Field("email", "REQUIRED", "email is required"),
		Field("email", "FORMAT", "email is malformed"),
		Field("age", "MIN", "too young"),
	)

	ce := ToConnectError(in)
	require.Equal(t, connect.CodeInvalidArgument, ce.Code())

	out := ToResultError(ce)
	require.Equal(t, CodeValidation, out.Code)
	require.Equal(t, http.StatusUnprocessableEntity, out.Status)
	require.Equal(t, in.Message, out.Message)
	require.Equal(t, in.Errors, out.Errors)
}

// 400 and 422 share a connect code; metadata keeps them apart.
func TestStatus_SurvivesSharedCode(t *testing.T) {
	bad := ToResultError(ToConnectError(BadRequestError("x")))
	invalid := ToResultError(ToConnectError(InvalidInputError("x")))

	require.Equal(t, http.StatusBadRequest, bad.Status)
	require.Equal(t, http.StatusUnprocessableEntity, invalid.Status)
}

func TestDomainCode_Preserved(t *testing.T) {
	in := NewResultErrorWithStatus("otp not found", "OTP_NOT_FOUND", http.StatusNotFound)

	out := ToResultError(ToConnectError(in))

	require.Equal(t, "OTP_NOT_FOUND", out.Code)
	require.Equal(t, http.StatusNotFound, out.Status)
	require.Empty(t, out.Errors)
}

// A server that doesn't use connectutil sends no ErrorInfo.
func TestToResultError_NoDetails(t *testing.T) {
	ce := connect.NewError(connect.CodeNotFound, errors.New("missing"))

	out := ToResultError(ce)

	require.Equal(t, "NOT_FOUND", out.Code)
	require.Equal(t, http.StatusNotFound, out.Status)
	require.Equal(t, "missing", out.Message)
	require.Nil(t, out.Errors)
}

func TestToConnectError_UnmappedStatus(t *testing.T) {
	in := NewResultErrorWithStatus("teapot", "TEAPOT", http.StatusTeapot)

	ce := ToConnectError(in)
	require.Equal(t, connect.CodeUnknown, ce.Code())
	require.Equal(t, http.StatusTeapot, ToResultError(ce).Status)
}
