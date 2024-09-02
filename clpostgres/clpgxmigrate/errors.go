package clpgxmigrate

import "fmt"

type RegisterError struct {
	filename string
	err      error
}

func registerError(filename string, msg string, args ...any) error {
	return fmt.Errorf("%w", RegisterError{filename: filename, err: fmt.Errorf(msg, args...)})
}

func (e RegisterError) Unwrap() error { return e.err }

func (e RegisterError) Error() string {
	return fmt.Sprintf("failed to register '%s': %v", e.filename, e.err)
}

type ApplyError struct {
	version int64
	err     error
}

func applyError(version int64, msg string, args ...any) error {
	return fmt.Errorf("%w", ApplyError{version: version, err: fmt.Errorf(msg, args...)})
}

func (e ApplyError) Unwrap() error {
	return e.err
}

func (e ApplyError) Error() string {
	return fmt.Sprintf("failed to apply migration version '%d': %v", e.version, e.err)
}
