package grpc

import (
	"errors"

	"github.com/nayefradwi/nayef_go_common/result"
)

var (
	GlobalWriterOnErrorListener result.OnErrorListener = func(err error) {}
)

type GrpcResponseWriter[T any] struct {
	ErrorListener result.OnErrorListener
}

func NewGrpcResponseWriter[T any]() GrpcResponseWriter[T] {
	return GrpcResponseWriter[T]{ErrorListener: GlobalWriterOnErrorListener}
}

func (gw GrpcResponseWriter[T]) WithErrorListener(listener result.OnErrorListener) GrpcResponseWriter[T] {
	gw.ErrorListener = listener
	return gw
}

func (gw GrpcResponseWriter[T]) WriteData(data T) (T, error) {
	return data, nil
}

func (gw GrpcResponseWriter[T]) WriteError(err error) error {
	gw.ErrorListener(err)

	var resultErr *result.ResultError
	if !errors.As(err, &resultErr) {
		resultErr = result.InternalError(err.Error())
	}

	return resultErr.ToGRPCError()
}

func (gw GrpcResponseWriter[T]) WriteResponse(data T, err error) (T, error) {
	if err != nil {
		return data, gw.WriteError(err)
	}

	return gw.WriteData(data)
}
