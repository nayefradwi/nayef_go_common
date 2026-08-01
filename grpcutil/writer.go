package grpcutil

import (
	"errors"
	"log/slog"

	. "github.com/nayefradwi/nayef_go_common/errors"
	"github.com/nayefradwi/nayef_go_common/errorspb"
)

var (
	GlobalWriterOnErrorListener OnErrorListener = func(err error) {
		slog.Error("GlobalWriterOnErrorListener", "error", err.Error())
	}
)

type GrpcResponseWriter[T any] struct {
	ErrorListener OnErrorListener
}

func NewGrpcResponseWriter[T any]() GrpcResponseWriter[T] {
	return GrpcResponseWriter[T]{ErrorListener: GlobalWriterOnErrorListener}
}

func (gw GrpcResponseWriter[T]) WithErrorListener(listener OnErrorListener) GrpcResponseWriter[T] {
	gw.ErrorListener = listener
	return gw
}

func (gw GrpcResponseWriter[T]) WriteData(data T) (T, error) {
	return data, nil
}

func (gw GrpcResponseWriter[T]) WriteError(err error) error {
	gw.ErrorListener(err)

	var resultErr *ResultError
	if !errors.As(err, &resultErr) {
		resultErr = InternalError("internal server error")
	}

	return errorspb.FromResultError(resultErr)
}

func (gw GrpcResponseWriter[T]) WriteResponse(data T, err error) (T, error) {
	if err != nil {
		return data, gw.WriteError(err)
	}

	return gw.WriteData(data)
}
