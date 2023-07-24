package climate

import "errors"

type usageError struct {
	error
}

func (uerr *usageError) Unwrap() error {
	return uerr.error
}

func ErrUsage(err error) *usageError {
	if err != nil {
		return &usageError{err}
	}
	return nil
}

type exitError struct {
	code int
	errs []error
}

func (eerr *exitError) Error() string {
	// We panic here if errs is empty, but this is somewhat intentional as we
	// should only use the error for exit code purposes in that case.
	return errors.Join(eerr.errs...).Error()
}

func (eerr *exitError) Unwrap() []error {
	return eerr.errs
}

func ErrExit(code int, errs ...error) *exitError {
	return &exitError{code, errs}
}
