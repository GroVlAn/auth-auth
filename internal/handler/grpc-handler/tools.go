package grpc_handler

import (
	"errors"

	"github.com/GroVlAn/auth-auth/internal/domain/e"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *GRPCHandler) handleError(err error) error {
	var errValidation *e.ErrValidation
	var errWrapper *e.ErrWrapper

	if errors.As(err, &errValidation) {
		field, reason, ok := errValidation.FirstError()

		if ok {
			h.l.Error().Err(errWrapper.Unwrap()).Msgf("validation error occurred: field: %s, reason: %s", field, reason)

			return status.Errorf(codes.InvalidArgument, "field: %s, error: %s", field, reason)
		}
	}

	if errors.As(err, &errWrapper) {
		return h.handleErrorWrapper(errWrapper)
	}

	return status.Error(codes.Internal, "internal server error")
}

func (h *GRPCHandler) handleErrorWrapper(errWrapper *e.ErrWrapper) error {
	switch errWrapper.ErrorType() {
	case e.ErrorTypeNotFound:
		h.l.Error().Err(errWrapper.Unwrap()).Msg("error not found occurred")

		return status.Error(codes.NotFound, errWrapper.Error())
	case e.ErrorTypeConflict:
		h.l.Error().Err(errWrapper.Unwrap()).Msg("error conflict occurred")

		return status.Error(codes.AlreadyExists, errWrapper.Error())
	case e.ErrorTypeUnauthorized:
		h.l.Error().Err(errWrapper.Unwrap()).Msg("error unauthorized occurred")

		return status.Error(codes.Unauthenticated, errWrapper.Error())
	case e.ErrorTypeInternal:
		h.l.Error().Err(errWrapper.Unwrap()).Msg("error internal occurred")

		return status.Error(codes.Internal, errWrapper.Error())
	default:
		h.l.Error().Err(errWrapper.Unwrap()).Msg("error internal(not wrapped) occurred")

		return status.Error(codes.Internal, errWrapper.Error())
	}
}

func (h *GRPCHandler) sendInternalError(logMessage string) error {
	h.l.Error().Msg(logMessage)

	return status.Error(codes.Internal, "internal server error")
}
