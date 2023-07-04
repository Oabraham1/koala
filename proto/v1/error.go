package proto

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	status "google.golang.org/grpc/status"
)

type ErrorOffsetOutOfRange struct {
	Offset uint64
}

func (error *ErrorOffsetOutOfRange) GRPCStatus() *status.Status {
	stat := status.New(400, fmt.Sprintf("Offset %d is out of range", error.Offset))
	message := fmt.Sprintf("Offset %d is out of range", error.Offset)
	d := &errdetails.LocalizedMessage{
		Locale:  "en-US",
		Message: message,
	}
	std, err := stat.WithDetails(d)
	if err != nil {
		return stat
	}
	return std
}

func (error *ErrorOffsetOutOfRange) Error() string {
	return error.GRPCStatus().Err().Error()
}
