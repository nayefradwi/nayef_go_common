package errorspb

import "google.golang.org/grpc/codes"

const (
	CodeCanceled           = "CANCELLED"
	CodeUnknown            = "UNKNOWN"
	CodeInvalidArgument    = "INVALID_ARGUMENT"
	CodeDeadlineExceeded   = "DEADLINE_EXCEEDED"
	CodeNotFound           = "NOT_FOUND"
	CodeAlreadyExists      = "ALREADY_EXISTS"
	CodePermissionDenied   = "PERMISSION_DENIED"
	CodeResourceExhausted  = "RESOURCE_EXHAUSTED"
	CodeFailedPrecondition = "FAILED_PRECONDITION"
	CodeAborted            = "ABORTED"
	CodeOutOfRange         = "OUT_OF_RANGE"
	CodeUnimplemented      = "UNIMPLEMENTED"
	CodeInternal           = "INTERNAL"
	CodeUnavailable        = "UNAVAILABLE"
	CodeDataLoss           = "DATA_LOSS"
	CodeUnauthenticated    = "UNAUTHENTICATED"
)

func newError(message, code string, grpcCode codes.Code) *ResultErrorPb {
	return &ResultErrorPb{Message: message, Code: code, Status: int32(grpcCode)}
}

func InvalidArgumentError(message string) *ResultErrorPb {
	return newError(message, CodeInvalidArgument, codes.InvalidArgument)
}

func NotFoundError(message string) *ResultErrorPb {
	return newError(message, CodeNotFound, codes.NotFound)
}

func AlreadyExistsError(message string) *ResultErrorPb {
	return newError(message, CodeAlreadyExists, codes.AlreadyExists)
}

func PermissionDeniedError(message string) *ResultErrorPb {
	return newError(message, CodePermissionDenied, codes.PermissionDenied)
}

func UnauthenticatedError(message string) *ResultErrorPb {
	return newError(message, CodeUnauthenticated, codes.Unauthenticated)
}

func FailedPreconditionError(message string) *ResultErrorPb {
	return newError(message, CodeFailedPrecondition, codes.FailedPrecondition)
}

func ResourceExhaustedError(message string) *ResultErrorPb {
	return newError(message, CodeResourceExhausted, codes.ResourceExhausted)
}

func AbortedError(message string) *ResultErrorPb {
	return newError(message, CodeAborted, codes.Aborted)
}

func OutOfRangeError(message string) *ResultErrorPb {
	return newError(message, CodeOutOfRange, codes.OutOfRange)
}

func DeadlineExceededError(message string) *ResultErrorPb {
	return newError(message, CodeDeadlineExceeded, codes.DeadlineExceeded)
}

func CanceledError(message string) *ResultErrorPb {
	return newError(message, CodeCanceled, codes.Canceled)
}

func UnimplementedError(message string) *ResultErrorPb {
	return newError(message, CodeUnimplemented, codes.Unimplemented)
}

func UnavailableError(message string) *ResultErrorPb {
	return newError(message, CodeUnavailable, codes.Unavailable)
}

func DataLossError(message string) *ResultErrorPb {
	return newError(message, CodeDataLoss, codes.DataLoss)
}

func InternalError(message string) *ResultErrorPb {
	return newError(message, CodeInternal, codes.Internal)
}

func UnknownError(message string) *ResultErrorPb {
	return newError(message, CodeUnknown, codes.Unknown)
}
