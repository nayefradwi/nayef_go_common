package connectutil

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	. "github.com/nayefradwi/nayef_go_common/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/proto"
)

// HTTP status rides along so 400 vs 422 survive the trip; both map to InvalidArgument.
const httpStatusKey = "httpStatus"

var httpToConnect = map[int]connect.Code{
	http.StatusBadRequest:          connect.CodeInvalidArgument,
	http.StatusUnauthorized:        connect.CodeUnauthenticated,
	http.StatusForbidden:           connect.CodePermissionDenied,
	http.StatusNotFound:            connect.CodeNotFound,
	http.StatusConflict:            connect.CodeAlreadyExists,
	http.StatusUnprocessableEntity: connect.CodeInvalidArgument,
	http.StatusTooManyRequests:     connect.CodeResourceExhausted,
	http.StatusInternalServerError: connect.CodeInternal,
	http.StatusNotImplemented:      connect.CodeUnimplemented,
	http.StatusServiceUnavailable:  connect.CodeUnavailable,
	http.StatusGatewayTimeout:      connect.CodeDeadlineExceeded,
}

var connectToHTTP = map[connect.Code]int{
	connect.CodeCanceled:           499,
	connect.CodeUnknown:            http.StatusInternalServerError,
	connect.CodeInvalidArgument:    http.StatusBadRequest,
	connect.CodeDeadlineExceeded:   http.StatusGatewayTimeout,
	connect.CodeNotFound:           http.StatusNotFound,
	connect.CodeAlreadyExists:      http.StatusConflict,
	connect.CodePermissionDenied:   http.StatusForbidden,
	connect.CodeUnauthenticated:    http.StatusUnauthorized,
	connect.CodeResourceExhausted:  http.StatusTooManyRequests,
	connect.CodeFailedPrecondition: http.StatusBadRequest,
	connect.CodeAborted:            http.StatusConflict,
	connect.CodeOutOfRange:         http.StatusBadRequest,
	connect.CodeUnimplemented:      http.StatusNotImplemented,
	connect.CodeInternal:           http.StatusInternalServerError,
	connect.CodeUnavailable:        http.StatusServiceUnavailable,
	connect.CodeDataLoss:           http.StatusInternalServerError,
}

func ToConnectError(e *ResultError) *connect.Error {
	if e == nil {
		return nil
	}

	status := statusOf(e)
	ce := connect.NewError(codeFromStatus(status), errors.New(e.Message))
	addDetail(ce, &errdetails.ErrorInfo{
		Reason:   e.Code,
		Metadata: map[string]string{httpStatusKey: strconv.Itoa(status)},
	})

	if len(e.Errors) > 0 {
		addDetail(ce, toBadRequest(e.Errors))
	}

	return ce
}

func ToResultError(ce *connect.Error) *ResultError {
	if ce == nil {
		return nil
	}

	result := &ResultError{
		Message: ce.Message(),
		Code:    strings.ToUpper(ce.Code().String()),
		Status:  statusFromCode(ce.Code()),
	}

	for _, d := range ce.Details() {
		msg, err := d.Value()
		if err != nil {
			continue
		}

		switch v := msg.(type) {
		case *errdetails.ErrorInfo:
			applyErrorInfo(result, v)
		case *errdetails.BadRequest:
			result.Errors = fromBadRequest(v)
		}
	}

	return result
}

func addDetail(ce *connect.Error, msg proto.Message) {
	detail, err := connect.NewErrorDetail(msg)
	if err != nil {
		return
	}
	ce.AddDetail(detail)
}

func applyErrorInfo(result *ResultError, info *errdetails.ErrorInfo) {
	if info.Reason != "" {
		result.Code = info.Reason
	}

	status, err := strconv.Atoi(info.Metadata[httpStatusKey])
	if err == nil {
		result.Status = status
	}
}

func toBadRequest(errs map[string][]ErrorDetails) *errdetails.BadRequest {
	fields := make([]string, 0, len(errs))
	for field := range errs {
		fields = append(fields, field)
	}
	slices.Sort(fields)

	br := &errdetails.BadRequest{}
	for _, field := range fields {
		for _, d := range errs[field] {
			br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
				Field:       field,
				Description: d.Message,
				Reason:      d.Code,
			})
		}
	}

	return br
}

func fromBadRequest(br *errdetails.BadRequest) map[string][]ErrorDetails {
	errs := make(map[string][]ErrorDetails, len(br.FieldViolations))
	for _, v := range br.FieldViolations {
		errs[v.Field] = append(errs[v.Field], ErrorDetails{
			Field:   v.Field,
			Message: v.Description,
			Code:    v.Reason,
		})
	}

	return errs
}

func codeFromStatus(status int) connect.Code {
	if code, ok := httpToConnect[status]; ok {
		return code
	}
	return connect.CodeUnknown
}

func statusFromCode(code connect.Code) int {
	if status, ok := connectToHTTP[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

func statusOf(e *ResultError) int {
	if e.Status >= 400 && e.Status <= 505 {
		return e.Status
	}

	switch e.Code {
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeUnknown, CodeInternal:
		return http.StatusInternalServerError
	case CodeInvalidInput, CodeValidation:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusBadRequest
	}
}
