package errors

import "fmt"

// Error is a base error that implement scw.SdkError
type Error struct {
	Str string
	Err error
}

// Error implement standard xerror.Wrapper interface
func (e *Error) Unwrap() error {
	return e.Err
}

// Error implement standard error interface
func (e *Error) Error() string {
	str := "scaleway-sdk-go: " + e.Str
	if e.Err != nil {
		str += ": " + e.Err.Error()
	}
	return str
}

// IsScwSdkError implement SdkError interface
func (e *Error) IsScwSdkError() {}

// New creates a new error with that same interface as fmt.Errorf
func New(format string, args ...interface{}) *Error {
	return &Error{
		Str: fmt.Sprintf(format, args...),
	}
}

// Wrap an error with additional information
func Wrap(err error, format string, args ...interface{}) *Error {
	return &Error{
		Err: err,
		Str: fmt.Sprintf(format, args...),
	}
}
