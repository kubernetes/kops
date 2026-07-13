package linodego

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strings"
)

const (
	ErrorUnsupported = iota
	// ErrorFromString is the Code identifying Errors created by string types
	ErrorFromString
	// ErrorFromError is the Code identifying Errors created by error types
	ErrorFromError
	// ErrorFromStringer is the Code identifying Errors created by fmt.Stringer types
	ErrorFromStringer
)

// Error wraps the LinodeGo error with the relevant http.Response
type Error struct {
	Response *http.Response
	Code     int
	Message  string
}

// APIErrorReason is an individual invalid request message returned by the Linode API
type APIErrorReason struct {
	Reason string `json:"reason"`
	Field  string `json:"field"`
}

func (r APIErrorReason) Error() string {
	if len(r.Field) == 0 {
		return r.Reason
	}

	return fmt.Sprintf("[%s] %s", r.Field, r.Reason)
}

// APIError is the error-set returned by the Linode API when presented with an invalid request
type APIError struct {
	Errors []APIErrorReason `json:"errors"`
}

//nolint:nestif,unparam
func coupleAPIErrors(resp *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return nil, NewError(err)
	}

	if resp == nil {
		return nil, NewError(fmt.Errorf("response is nil"))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Check that response is of the correct content-type before unmarshalling
		expectedContentType := ""
		if resp.Request != nil && resp.Request.Header != nil {
			expectedContentType = resp.Request.Header.Get("Accept")
		}

		responseContentType := resp.Header.Get("Content-Type")

		// If the upstream server fails to respond to the request,
		// the HTTP server will respond with a default error page with Content-Type "text/html".
		if resp.StatusCode == http.StatusBadGateway && responseContentType == "text/html" {
			return nil, &Error{Code: http.StatusBadGateway, Message: http.StatusText(http.StatusBadGateway), Response: resp}
		}

		if responseContentType != expectedContentType {
			if resp.Body == nil {
				return nil, NewError(fmt.Errorf("response body is nil"))
			}

			bodyBytes, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				return nil, NewError(fmt.Errorf("failed to read response body: %w", readErr))
			}

			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			msg := fmt.Sprintf(
				"Unexpected Content-Type: Expected: %v, Received: %v\nResponse body: %s",
				expectedContentType,
				responseContentType,
				string(bodyBytes),
			)

			return nil, &Error{Code: resp.StatusCode, Message: msg, Response: resp}
		}

		// Must check if there is no list of reasons in the error before making a call to NewError
		apiError, ok := getAPIError(resp)
		if !ok {
			return nil, NewError(fmt.Errorf("failed to decode response body"))
		}

		if len(apiError.Errors) == 0 {
			return resp, nil
		}

		return nil, NewError(resp)
	}

	return resp, nil
}

func (e APIError) Error() string {
	x := make([]string, 0, len(e.Errors))
	for _, msg := range e.Errors {
		x = append(x, msg.Error())
	}

	return strings.Join(x, "; ")
}

// NewError creates a linodego.Error with a Code identifying the source err type,
// - ErrorFromString   (1) from a string
// - ErrorFromError    (2) for an error
// - ErrorFromStringer (3) for a Stringer
// - HTTP Status Codes (100-600) for a http.Response object
func NewError(err any) *Error {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case *Error:
		return e
	case *http.Response:
		apiError, ok := getAPIError(e)

		if !ok {
			return &Error{Code: ErrorUnsupported, Message: "Unexpected HTTP Error Response, no error"}
		}

		return &Error{
			Code:     e.StatusCode,
			Message:  apiError.Error(),
			Response: e,
		}
	case error:
		return &Error{Code: ErrorFromError, Message: e.Error()}
	case string:
		return &Error{Code: ErrorFromString, Message: e}
	case fmt.Stringer:
		return &Error{Code: ErrorFromStringer, Message: e.String()}
	default:
		return &Error{Code: ErrorUnsupported, Message: fmt.Sprintf("Unsupported type to linodego.NewError: %s", reflect.TypeOf(e))}
	}
}

func (err Error) Error() string {
	return fmt.Sprintf("[%03d] %s", err.Code, err.Message)
}

func (err Error) StatusCode() int {
	return err.Code
}

func (err Error) Is(target error) bool {
	if x, ok := target.(interface{ StatusCode() int }); ok || errors.As(target, &x) {
		return err.StatusCode() == x.StatusCode()
	}

	return false
}

// IsNotFound indicates if err indicates a 404 Not Found error from the Linode API.
func IsNotFound(err error) bool {
	return ErrHasStatus(err, http.StatusNotFound)
}

// ErrHasStatus checks if err is an error from the Linode API, and whether it contains the given HTTP status code.
// More than one status code may be given.
// If len(code) == 0, err is nil or is not a [Error], ErrHasStatus will return false.
func ErrHasStatus(err error, code ...int) bool {
	if err == nil {
		return false
	}

	// Short-circuit if the caller did not provide any status codes.
	if len(code) == 0 {
		return false
	}

	var e *Error
	if !errors.As(err, &e) {
		return false
	}

	ec := e.StatusCode()

	return slices.Contains(code, ec)
}
