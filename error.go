package climate

import "errors"

type usageError struct {
	error
}

// ErrUsage returns the given error wrapped in a usageError or nil otherwise.
// usageError is used to indicate there's something wrong with the user input
// and that the usage information should be printed along with the error.
func ErrUsage(err error) *usageError {
	if err != nil {
		return &usageError{err}
	}
	return nil
}

func (uerr *usageError) Unwrap() error {
	return uerr.error
}

type exitError struct {
	code int
	errs []error
}

// ErrExit returns an exitError with the given exit code and errors.
// exitError is used to indicate that the CLI should exit with the given exit
// code (as returned by Run and respected by RunAndExit).
func ErrExit(code int, errs ...error) *exitError {
	return &exitError{code, errs}
}

func (eerr *exitError) Error() string {
	// We panic here if errs is empty, but this is somewhat intentional as we
	// should only use the error for exit code purposes in that case.
	return errors.Join(eerr.errs...).Error()
}

func (eerr *exitError) Unwrap() []error {
	return eerr.errs
}
