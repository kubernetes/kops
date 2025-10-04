package services

import (
	"errors"

	"github.com/aws/smithy-go"
)

// MockAPIError mocks smithy.APIError
type MockAPIError struct {
	error
	code    string
	message string
}

// NewMockAPIError returns a new APIError
func NewMockAPIError(code string, message string) smithy.APIError {
	return &MockAPIError{
		error:   errors.New(message),
		code:    code,
		message: message,
	}
}

// ErrorCode returns the error code
func (e *MockAPIError) ErrorCode() string {
	return e.code
}

// ErrorMessage returns the error message
func (e *MockAPIError) ErrorMessage() string {
	return e.message
}

// ErrorFault isn't really implemented.
func (e *MockAPIError) ErrorFault() smithy.ErrorFault {
	return 1
}
