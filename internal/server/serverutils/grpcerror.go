package serverutils

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/goto/entropy/pkg/errors"
)

// ToRPCError returns an instance of gRPC Error Status equivalent to the
// given error value.
func ToRPCError(e error) error {
	err := errors.E(e)

	var code codes.Code
	switch {
	case errors.Is(err, errors.ErrNotFound):
		code = codes.NotFound

	case errors.Is(err, errors.ErrConflict):
		code = codes.AlreadyExists

	case errors.Is(err, errors.ErrInvalid):
		code = codes.InvalidArgument

	default:
		code = codes.Internal
	}
	return status.Error(code, err.Error())
}
