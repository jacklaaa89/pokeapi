package helpers

import (
	"context"
	"net/http"

	"github.com/jacklaaa89/pokeapi/internal/api/errors"
	"github.com/jacklaaa89/pokeapi/internal/server/middleware"
)

// compoundError represents an error with more context.
type compoundError interface {
	error
	Code() errors.Code // Code returns the error code to apply.
	StatusCode() int   // StatusCode returns the status code to apply.
}

// errorResponse the response returned from our API in the result of an error.
type errorResponse struct {
	Error string      `json:"error"`          // Error is the error message
	Code  errors.Code `json:"code,omitempty"` // Code is any optional error code.
}

type invalidRequestError struct{ err error }

func (i *invalidRequestError) Error() string   { return i.err.Error() }
func (*invalidRequestError) Code() errors.Code { return errors.CodeInvalidRequest }
func (*invalidRequestError) StatusCode() int   { return http.StatusBadRequest }

// InvalidRequest wraps an error to return http.StatusBadRequest as the status code
// and errors.CodeInvalidRequest as the error code.
func InvalidRequest(err error) error { return &invalidRequestError{err} }

// RespondError this allows us to write an error to the supplied http.ResponseWriter
func RespondError(ctx context.Context, w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	var (
		status = http.StatusInternalServerError
		msg    = err.Error()
		code   = errors.CodeServerError
	)

	switch v := err.(type) {
	case *errors.Error:
		if v.StatusCode > 0 {
			status = v.StatusCode
		}
		code = v.Code
		msg = string(v.Code)
	case compoundError:
		status = v.StatusCode()
		code = v.Code()
	}

	res := &errorResponse{Error: msg, Code: code}
	write(ctx, w, status, &response{RequestID: middleware.RequestID(ctx), Error: res})
}
