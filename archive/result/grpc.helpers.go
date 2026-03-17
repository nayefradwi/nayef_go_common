package result

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (e *ResultError) ToGRPCError() error {
	st := status.New(codeToGRPCCode(e.Code), e.Message)
	st, _ = st.WithDetails(e.ToProto())
	return st.Err()
}

func FromGRPCError(err error) *ResultError {
	st, ok := status.FromError(err)
	if !ok {
		return &ResultError{Message: err.Error(), Code: UNKNOWN_ERROR_CODE}
	}

	for _, d := range st.Details() {
		if pbErr, ok := d.(*ResultErrorPb); ok {
			return FromProto(pbErr)
		}
	}

	return &ResultError{Message: st.Message(), Code: grpcCodeToCode(st.Code())}
}

func codeToGRPCCode(code string) codes.Code {
	switch code {
	case NOT_FOUND_CODE:
		return codes.NotFound
	case BAD_REQUEST_CODE, VALIDATION_ERROR_CODE, INVALID_INPUT_CODE:
		return codes.InvalidArgument
	case UNAUTHORIZED_CODE:
		return codes.Unauthenticated
	case FORBIDDEN_CODE:
		return codes.PermissionDenied
	default:
		return codes.Internal
	}
}

func grpcCodeToCode(c codes.Code) string {
	switch c {
	case codes.NotFound:
		return NOT_FOUND_CODE
	case codes.InvalidArgument:
		return BAD_REQUEST_CODE
	case codes.Unauthenticated:
		return UNAUTHORIZED_CODE
	case codes.PermissionDenied:
		return FORBIDDEN_CODE
	default:
		return INTERNAL_ERROR_CODE
	}
}
