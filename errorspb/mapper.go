package errorspb

import (
	"net/http"

	customerrors "github.com/nayefradwi/nayef_go_common/errors"
	"google.golang.org/grpc/codes"
)

var grpcToHTTP = map[codes.Code]int{
	codes.OK:                 http.StatusOK,
	codes.Canceled:           499,
	codes.Unknown:            http.StatusInternalServerError,
	codes.InvalidArgument:    http.StatusBadRequest,
	codes.DeadlineExceeded:   http.StatusGatewayTimeout,
	codes.NotFound:           http.StatusNotFound,
	codes.AlreadyExists:      http.StatusConflict,
	codes.PermissionDenied:   http.StatusForbidden,
	codes.Unauthenticated:    http.StatusUnauthorized,
	codes.ResourceExhausted:  http.StatusTooManyRequests,
	codes.FailedPrecondition: http.StatusBadRequest,
	codes.Aborted:            http.StatusConflict,
	codes.OutOfRange:         http.StatusBadRequest,
	codes.Unimplemented:      http.StatusNotImplemented,
	codes.Internal:           http.StatusInternalServerError,
	codes.Unavailable:        http.StatusServiceUnavailable,
	codes.DataLoss:           http.StatusInternalServerError,
}

var httpToGRPC = map[int]codes.Code{
	http.StatusBadRequest:          codes.InvalidArgument,
	http.StatusUnauthorized:        codes.Unauthenticated,
	http.StatusForbidden:           codes.PermissionDenied,
	http.StatusNotFound:            codes.NotFound,
	http.StatusConflict:            codes.AlreadyExists,
	http.StatusUnprocessableEntity: codes.InvalidArgument,
	http.StatusTooManyRequests:     codes.ResourceExhausted,
	http.StatusInternalServerError: codes.Internal,
	http.StatusNotImplemented:      codes.Unimplemented,
	http.StatusServiceUnavailable:  codes.Unavailable,
	http.StatusGatewayTimeout:      codes.DeadlineExceeded,
}

func HTTPStatusFromGRPCCode(c codes.Code) int {
	if s, ok := grpcToHTTP[c]; ok {
		return s
	}
	return http.StatusInternalServerError
}

func GRPCCodeFromHTTPStatus(status int) codes.Code {
	if c, ok := httpToGRPC[status]; ok {
		return c
	}
	return codes.Unknown
}

func FromResultError(e *customerrors.ResultError) *ResultErrorPb {
	if e == nil {
		return nil
	}
	pbErrors := make(map[string]*ErrorDetailsPbList, len(e.Errors))
	for field, details := range e.Errors {
		items := make([]*ErrorDetailsPb, len(details))
		for i, d := range details {
			items[i] = &ErrorDetailsPb{
				Message: d.Message,
				Code:    d.Code,
				Field:   d.Field,
			}
		}
		pbErrors[field] = &ErrorDetailsPbList{Items: items}
	}
	return &ResultErrorPb{
		Message: e.Message,
		Code:    e.Code,
		Status:  int32(GRPCCodeFromHTTPStatus(e.Status)),
		Errors:  pbErrors,
	}
}

func ToResultError(pb *ResultErrorPb) *customerrors.ResultError {
	if pb == nil {
		return nil
	}
	errs := make(map[string][]customerrors.ErrorDetails, len(pb.Errors))
	for field, list := range pb.Errors {
		if list == nil {
			continue
		}
		details := make([]customerrors.ErrorDetails, len(list.Items))
		for i, item := range list.Items {
			details[i] = customerrors.ErrorDetails{
				Message: item.Message,
				Code:    item.Code,
				Field:   item.Field,
			}
		}
		errs[field] = details
	}
	return &customerrors.ResultError{
		Message: pb.Message,
		Code:    pb.Code,
		Errors:  errs,
		Status:  HTTPStatusFromGRPCCode(codes.Code(pb.Status)),
	}
}
