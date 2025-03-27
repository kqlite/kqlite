package pgerror

import (
	"errors"
)

type errWithCode struct {
	cause error
	code  string
}

var _ error = (*errWithCode)(nil)

func (erc *errWithCode) Error() string { return erc.cause.Error() }

// ErrorWithCode decorates the error with postgres error code.
// that can be fetched by GetPGCode() below conditionally.
func ErrWithCode(err error, code string) error {
	if err == nil {
		return nil
	}
	return &errWithCode{cause: err, code: code}
}

// New creates an error with a code.
func New(code, msg string) error {
	err := errors.New(msg)
	err = ErrWithCode(err, code)
	return err
}

// GetPGCode retrieves the error PostgreSQL code for an error if present.
func GetPGCode(err error) string {
	code := ""
	if erc, ok := err.(*errWithCode); ok {
		code = erc.code
	}

	return code
}
