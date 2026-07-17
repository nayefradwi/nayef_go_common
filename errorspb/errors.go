package errorspb

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (err *ResultErrorPb) Error() string {
	return err.Message
}

func (err *ResultErrorPb) GRPCStatus() *status.Status {
	st := status.New(codes.Code(err.Status), err.Message)
	if withDetails, detailErr := st.WithDetails(err); detailErr == nil {
		return withDetails
	}
	return st
}
