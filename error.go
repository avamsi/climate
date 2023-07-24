package climate

import (
	"errors"
	"fmt"
)

type usageError struct {
	err error
}

func (uerr *usageError) Error() string {
	return uerr.err.Error()
}

func (uerr *usageError) Unwrap() error {
	return uerr.err
}

func ErrUsage(err error) *usageError {
	if err == nil { // if _no_ error
		return nil
	}
	return &usageError{err}
}

type exitError struct {
	code int
	errs []error
}

func (eerr *exitError) Error() string {
	var text string
	if err := errors.Join(eerr.errs...); err != nil {
		text = err.Error()
	}
	return fmt.Sprintf("%d: %s", eerr.code, text)
}

func (eerr *exitError) Unwrap() []error {
	return eerr.errs
}

func ErrExit(code int, errs ...error) *exitError {
	return &exitError{code, errs}
}
