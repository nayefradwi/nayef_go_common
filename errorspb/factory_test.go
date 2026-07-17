package errorspb

import (
	"net/http"
	"testing"

	customerrors "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFactories(t *testing.T) {
	cases := []struct {
		name       string
		err        *ResultErrorPb
		wantCode   string
		wantStatus codes.Code
	}{
		{"InvalidArgument", InvalidArgumentError("m"), CodeInvalidArgument, codes.InvalidArgument},
		{"NotFound", NotFoundError("m"), CodeNotFound, codes.NotFound},
		{"AlreadyExists", AlreadyExistsError("m"), CodeAlreadyExists, codes.AlreadyExists},
		{"PermissionDenied", PermissionDeniedError("m"), CodePermissionDenied, codes.PermissionDenied},
		{"Unauthenticated", UnauthenticatedError("m"), CodeUnauthenticated, codes.Unauthenticated},
		{"FailedPrecondition", FailedPreconditionError("m"), CodeFailedPrecondition, codes.FailedPrecondition},
		{"ResourceExhausted", ResourceExhaustedError("m"), CodeResourceExhausted, codes.ResourceExhausted},
		{"Aborted", AbortedError("m"), CodeAborted, codes.Aborted},
		{"OutOfRange", OutOfRangeError("m"), CodeOutOfRange, codes.OutOfRange},
		{"DeadlineExceeded", DeadlineExceededError("m"), CodeDeadlineExceeded, codes.DeadlineExceeded},
		{"Canceled", CanceledError("m"), CodeCanceled, codes.Canceled},
		{"Unimplemented", UnimplementedError("m"), CodeUnimplemented, codes.Unimplemented},
		{"Unavailable", UnavailableError("m"), CodeUnavailable, codes.Unavailable},
		{"DataLoss", DataLossError("m"), CodeDataLoss, codes.DataLoss},
		{"Internal", InternalError("m"), CodeInternal, codes.Internal},
		{"Unknown", UnknownError("m"), CodeUnknown, codes.Unknown},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, "m", tc.err.Message)
			require.Equal(t, tc.wantCode, tc.err.Code)
			require.Equal(t, int32(tc.wantStatus), tc.err.Status)
		})
	}
}

func TestGRPCStatus(t *testing.T) {
	err := NotFoundError("missing")
	st := status.Convert(err)
	require.Equal(t, codes.NotFound, st.Code())
	require.Equal(t, "missing", st.Message())

	err2 := InvalidArgumentError("bad")
	require.Equal(t, codes.InvalidArgument, status.Code(err2))
}

func TestHTTPStatusFromGRPCCode(t *testing.T) {
	require.Equal(t, http.StatusNotFound, HTTPStatusFromGRPCCode(codes.NotFound))
	require.Equal(t, http.StatusBadRequest, HTTPStatusFromGRPCCode(codes.InvalidArgument))
	require.Equal(t, http.StatusInternalServerError, HTTPStatusFromGRPCCode(codes.Code(999)))
}

func TestGRPCCodeFromHTTPStatus(t *testing.T) {
	require.Equal(t, codes.NotFound, GRPCCodeFromHTTPStatus(http.StatusNotFound))
	require.Equal(t, codes.InvalidArgument, GRPCCodeFromHTTPStatus(http.StatusUnprocessableEntity))
	require.Equal(t, codes.Unknown, GRPCCodeFromHTTPStatus(0))
}

func TestMapperStatusTranslation(t *testing.T) {
	pb := FromResultError(customerrors.NotFoundError("x"))
	require.Equal(t, int32(codes.NotFound), pb.Status)

	back := ToResultError(NotFoundError("x"))
	require.Equal(t, http.StatusNotFound, back.Status)
}
