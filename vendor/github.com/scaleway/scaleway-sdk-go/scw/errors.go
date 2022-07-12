package scw

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/validation"
)

// SdkError is a base interface for all Scaleway SDK errors.
type SdkError interface {
	Error() string
	IsScwSdkError()
}

// ResponseError is an error type for the Scaleway API
type ResponseError struct {
	// Message is a human-friendly error message
	Message string `json:"message"`

	// Type is a string code that defines the kind of error. This field is only used by instance API
	Type string `json:"type,omitempty"`

	// Resource is a string code that defines the resource concerned by the error. This field is only used by instance API
	Resource string `json:"resource,omitempty"`

	// Fields contains detail about validation error. This field is only used by instance API
	Fields map[string][]string `json:"fields,omitempty"`

	// StatusCode is the HTTP status code received
	StatusCode int `json:"-"`

	// Status is the HTTP status received
	Status string `json:"-"`

	RawBody json.RawMessage `json:"-"`
}

func (e *ResponseError) UnmarshalJSON(b []byte) error {
	type tmpResponseError ResponseError
	tmp := tmpResponseError(*e)

	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	tmp.Message = strings.ToLower(tmp.Message)

	*e = ResponseError(tmp)
	return nil
}

// IsScwSdkError implement SdkError interface
func (e *ResponseError) IsScwSdkError() {}
func (e *ResponseError) Error() string {
	s := fmt.Sprintf("scaleway-sdk-go: http error %s", e.Status)

	if e.Resource != "" {
		s = fmt.Sprintf("%s: resource %s", s, e.Resource)
	}

	if e.Message != "" {
		s = fmt.Sprintf("%s: %s", s, e.Message)
	}

	if len(e.Fields) > 0 {
		s = fmt.Sprintf("%s: %v", s, e.Fields)
	}

	return s
}
func (e *ResponseError) GetRawBody() json.RawMessage {
	return e.RawBody
}

// hasResponseError returns an SdkError when the HTTP status is not OK.
func hasResponseError(res *http.Response) error {
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}

	newErr := &ResponseError{
		StatusCode: res.StatusCode,
		Status:     res.Status,
	}

	if res.Body == nil {
		return newErr
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "cannot read error response body")
	}
	newErr.RawBody = body

	// The error content is not encoded in JSON, only returns HTTP data.
	if res.Header.Get("Content-Type") != "application/json" {
		newErr.Message = res.Status
		return newErr
	}

	err = json.Unmarshal(body, newErr)
	if err != nil {
		return errors.Wrap(err, "could not parse error response body")
	}

	err = unmarshalStandardError(newErr.Type, body)
	if err != nil {
		return err
	}

	err = unmarshalNonStandardError(newErr.Type, body)
	if err != nil {
		return err
	}

	return newErr
}

func unmarshalStandardError(errorType string, body []byte) error {
	var stdErr SdkError

	switch errorType {
	case "invalid_arguments":
		stdErr = &InvalidArgumentsError{RawBody: body}
	case "quotas_exceeded":
		stdErr = &QuotasExceededError{RawBody: body}
	case "transient_state":
		stdErr = &TransientStateError{RawBody: body}
	case "not_found":
		stdErr = &ResourceNotFoundError{RawBody: body}
	case "locked":
		stdErr = &ResourceLockedError{RawBody: body}
	case "permissions_denied":
		stdErr = &PermissionsDeniedError{RawBody: body}
	case "out_of_stock":
		stdErr = &OutOfStockError{RawBody: body}
	case "resource_expired":
		stdErr = &ResourceExpiredError{RawBody: body}
	case "denied_authentication":
		stdErr = &DeniedAuthenticationError{RawBody: body}
	case "precondition_failed":
		stdErr = &PreconditionFailedError{RawBody: body}
	default:
		return nil
	}

	err := json.Unmarshal(body, stdErr)
	if err != nil {
		return errors.Wrap(err, "could not parse error %s response body", errorType)
	}

	return stdErr
}

func unmarshalNonStandardError(errorType string, body []byte) error {
	switch errorType {
	// Only in instance API.

	case "unknown_resource":
		unknownResourceError := &UnknownResource{RawBody: body}
		err := json.Unmarshal(body, unknownResourceError)
		if err != nil {
			return errors.Wrap(err, "could not parse error %s response body", errorType)
		}
		return unknownResourceError.ToResourceNotFoundError()

	case "invalid_request_error":
		invalidRequestError := &InvalidRequestError{RawBody: body}
		err := json.Unmarshal(body, invalidRequestError)
		if err != nil {
			return errors.Wrap(err, "could not parse error %s response body", errorType)
		}

		invalidArgumentsError := invalidRequestError.ToInvalidArgumentsError()
		if invalidArgumentsError != nil {
			return invalidArgumentsError
		}

		quotasExceededError := invalidRequestError.ToQuotasExceededError()
		if quotasExceededError != nil {
			return quotasExceededError
		}

		// At this point, the invalid_request_error is not an InvalidArgumentsError and
		// the default marshalling will be used.
		return nil

	default:
		return nil
	}
}

type InvalidArgumentsErrorDetail struct {
	ArgumentName string `json:"argument_name"`
	Reason       string `json:"reason"`
	HelpMessage  string `json:"help_message"`
}

type InvalidArgumentsError struct {
	Details []InvalidArgumentsErrorDetail `json:"details"`

	RawBody json.RawMessage `json:"-"`
}

// IsScwSdkError implements the SdkError interface
func (e *InvalidArgumentsError) IsScwSdkError() {}
func (e *InvalidArgumentsError) Error() string {
	invalidArgs := make([]string, len(e.Details))
	for i, d := range e.Details {
		invalidArgs[i] = d.ArgumentName
		switch d.Reason {
		case "unknown":
			invalidArgs[i] += " is invalid for unexpected reason"
		case "required":
			invalidArgs[i] += " is required"
		case "format":
			invalidArgs[i] += " is wrongly formatted"
		case "constraint":
			invalidArgs[i] += " does not respect constraint"
		}
		if d.HelpMessage != "" {
			invalidArgs[i] += ", " + d.HelpMessage
		}
	}

	return "scaleway-sdk-go: invalid argument(s): " + strings.Join(invalidArgs, "; ")
}
func (e *InvalidArgumentsError) GetRawBody() json.RawMessage {
	return e.RawBody
}

// UnknownResource is only returned by the instance API.
// Warning: this is not a standard error.
type UnknownResource struct {
	Message string          `json:"message"`
	RawBody json.RawMessage `json:"-"`
}

// ToSdkError returns a standard error InvalidArgumentsError or nil Fields is nil.
func (e *UnknownResource) ToResourceNotFoundError() SdkError {
	resourceNotFound := &ResourceNotFoundError{
		RawBody: e.RawBody,
	}

	messageParts := strings.Split(e.Message, `"`)

	// Some errors uses ' and not "
	if len(messageParts) == 1 {
		messageParts = strings.Split(e.Message, "'")
	}

	switch len(messageParts) {
	case 2: // message like: `"111..." not found`
		resourceNotFound.ResourceID = messageParts[0]
	case 3: // message like: `Security Group "111..." not found`
		resourceNotFound.ResourceID = messageParts[1]
		// transform `Security group ` to `security_group`
		resourceNotFound.Resource = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(messageParts[0])), " ", "_")
	default:
		return nil
	}
	if !validation.IsUUID(resourceNotFound.ResourceID) {
		return nil
	}
	return resourceNotFound
}

// InvalidRequestError is only returned by the instance API.
// Warning: this is not a standard error.
type InvalidRequestError struct {
	Message string `json:"message"`

	Fields map[string][]string `json:"fields"`

	Resource string `json:"resource"`

	RawBody json.RawMessage `json:"-"`
}

// ToSdkError returns a standard error InvalidArgumentsError or nil Fields is nil.
func (e *InvalidRequestError) ToInvalidArgumentsError() SdkError {
	// If error has no fields, it is not an InvalidArgumentsError.
	if e.Fields == nil || len(e.Fields) == 0 {
		return nil
	}

	invalidArguments := &InvalidArgumentsError{
		RawBody: e.RawBody,
	}
	fieldNames := []string(nil)
	for fieldName := range e.Fields {
		fieldNames = append(fieldNames, fieldName)
	}
	sort.Strings(fieldNames)
	for _, fieldName := range fieldNames {
		for _, message := range e.Fields[fieldName] {
			invalidArguments.Details = append(invalidArguments.Details, InvalidArgumentsErrorDetail{
				ArgumentName: fieldName,
				Reason:       "constraint",
				HelpMessage:  message,
			})
		}
	}
	return invalidArguments
}

func (e *InvalidRequestError) ToQuotasExceededError() SdkError {
	if !strings.Contains(strings.ToLower(e.Message), "quota exceeded for this resource") {
		return nil
	}

	return &QuotasExceededError{
		Details: []QuotasExceededErrorDetail{
			{
				Resource: e.Resource,
				Quota:    0,
				Current:  0,
			},
		},
		RawBody: e.RawBody,
	}
}

type QuotasExceededErrorDetail struct {
	Resource string `json:"resource"`
	Quota    uint32 `json:"quota"`
	Current  uint32 `json:"current"`
}

type QuotasExceededError struct {
	Details []QuotasExceededErrorDetail `json:"details"`
	RawBody json.RawMessage             `json:"-"`
}

// IsScwSdkError implements the SdkError interface
func (e *QuotasExceededError) IsScwSdkError() {}
func (e *QuotasExceededError) Error() string {
	invalidArgs := make([]string, len(e.Details))
	for i, d := range e.Details {
		invalidArgs[i] = fmt.Sprintf("%s has reached its quota (%d/%d)", d.Resource, d.Current, d.Current)
	}

	return "scaleway-sdk-go: quota exceeded(s): " + strings.Join(invalidArgs, "; ")
}
func (e *QuotasExceededError) GetRawBody() json.RawMessage {
	return e.RawBody
}

type PermissionsDeniedError struct {
	Details []struct {
		Resource string `json:"resource"`
		Action   string `json:"action"`
	} `json:"details"`

	RawBody json.RawMessage `json:"-"`
}

// IsScwSdkError implements the SdkError interface
func (e *PermissionsDeniedError) IsScwSdkError() {}
func (e *PermissionsDeniedError) Error() string {
	invalidArgs := make([]string, len(e.Details))
	for i, d := range e.Details {
		invalidArgs[i] = fmt.Sprintf("%s %s", d.Action, d.Resource)
	}

	return "scaleway-sdk-go: insufficient permissions: " + strings.Join(invalidArgs, "; ")
}
func (e *PermissionsDeniedError) GetRawBody() json.RawMessage {
	return e.RawBody
}

type TransientStateError struct {
	Resource     string `json:"resource"`
	ResourceID   string `json:"resource_id"`
	CurrentState string `json:"current_state"`

	RawBody json.RawMessage `json:"-"`
}

// IsScwSdkError implements the SdkError interface
func (e *TransientStateError) IsScwSdkError() {}
func (e *TransientStateError) Error() string {
	return fmt.Sprintf("scaleway-sdk-go: resource %s with ID %s is in a transient state: %s", e.Resource, e.ResourceID, e.CurrentState)
}
func (e *TransientStateError) GetRawBody() json.RawMessage {
	return e.RawBody
}

type ResourceNotFoundError struct {
	Resource   string `json:"resource"`
	ResourceID string `json:"resource_id"`

	RawBody json.RawMessage `json:"-"`
}

// IsScwSdkError implements the SdkError interface
func (e *ResourceNotFoundError) IsScwSdkError() {}
func (e *ResourceNotFoundError) Error() string {
	return fmt.Sprintf("scaleway-sdk-go: resource %s with ID %s is not found", e.Resource, e.ResourceID)
}
func (e *ResourceNotFoundError) GetRawBody() json.RawMessage {
	return e.RawBody
}

type ResourceLockedError struct {
	Resource   string `json:"resource"`
	ResourceID string `json:"resource_id"`

	RawBody json.RawMessage `json:"-"`
}

// IsScwSdkError implements the SdkError interface
func (e *ResourceLockedError) IsScwSdkError() {}
func (e *ResourceLockedError) Error() string {
	return fmt.Sprintf("scaleway-sdk-go: resource %s with ID %s is locked", e.Resource, e.ResourceID)
}
func (e *ResourceLockedError) GetRawBody() json.RawMessage {
	return e.RawBody
}

type OutOfStockError struct {
	Resource string `json:"resource"`

	RawBody json.RawMessage `json:"-"`
}

// IsScwSdkError implements the SdkError interface
func (e *OutOfStockError) IsScwSdkError() {}
func (e *OutOfStockError) Error() string {
	return fmt.Sprintf("scaleway-sdk-go: resource %s is out of stock", e.Resource)
}
func (e *OutOfStockError) GetRawBody() json.RawMessage {
	return e.RawBody
}

// InvalidClientOptionError indicates that at least one of client data has been badly provided for the client creation.
type InvalidClientOptionError struct {
	errorType string
}

func NewInvalidClientOptionError(format string, a ...interface{}) *InvalidClientOptionError {
	return &InvalidClientOptionError{errorType: fmt.Sprintf(format, a...)}
}

// IsScwSdkError implements the SdkError interface
func (e InvalidClientOptionError) IsScwSdkError() {}
func (e InvalidClientOptionError) Error() string {
	return fmt.Sprintf("scaleway-sdk-go: %s", e.errorType)
}

// ConfigFileNotFound indicates that the config file could not be found
type ConfigFileNotFoundError struct {
	path string
}

func configFileNotFound(path string) *ConfigFileNotFoundError {
	return &ConfigFileNotFoundError{path: path}
}

// ConfigFileNotFoundError implements the SdkError interface
func (e ConfigFileNotFoundError) IsScwSdkError() {}
func (e ConfigFileNotFoundError) Error() string {
	return fmt.Sprintf("scaleway-sdk-go: cannot read config file %s: no such file or directory", e.path)
}

// ResourceExpiredError implements the SdkError interface
type ResourceExpiredError struct {
	Resource     string    `json:"resource"`
	ResourceID   string    `json:"resource_id"`
	ExpiredSince time.Time `json:"expired_since"`

	RawBody json.RawMessage `json:"-"`
}

func (r ResourceExpiredError) Error() string {
	return fmt.Sprintf("scaleway-sdk-go: resource %s with ID %s expired since %s", r.Resource, r.ResourceID, r.ExpiredSince.String())
}

func (r ResourceExpiredError) IsScwSdkError() {}

// DeniedAuthenticationError implements the SdkError interface
type DeniedAuthenticationError struct {
	Method string `json:"method"`
	Reason string `json:"reason"`

	RawBody json.RawMessage `json:"-"`
}

func (r DeniedAuthenticationError) Error() string {
	var reason string
	var method string

	switch r.Method {
	case "unknown_method":
		method = "unknown method"
	case "jwt":
		method = "JWT"
	case "api_key":
		method = "API key"
	}

	switch r.Reason {
	case "unknown_reason":
		reason = "unknown reason"
	case "invalid_argument":
		reason = "invalid " + method + " format or empty value"
	case "not_found":
		reason = method + " does not exist"
	case "expired":
		reason = method + " is expired"
	}
	return fmt.Sprintf("scaleway-sdk-go: denied authentication: %s", reason)
}

func (r DeniedAuthenticationError) IsScwSdkError() {}

// PreconditionFailedError implements the SdkError interface
type PreconditionFailedError struct {
	Precondition string `json:"method"`
	HelpMessage  string `json:"help_message"`

	RawBody json.RawMessage `json:"-"`
}

func (r PreconditionFailedError) Error() string {
	var msg string
	switch r.Precondition {
	case "unknown_precondition":
		msg = "unknown precondition"
	case "resource_still_in_use":
		msg = "resource is still in use"
	case "attribute_must_be_set":
		msg = "attribute must be set"
	}
	if r.HelpMessage != "" {
		msg += ", " + r.HelpMessage
	}

	return fmt.Sprintf("scaleway-sdk-go: precondition failed: %s", msg)
}

func (r PreconditionFailedError) IsScwSdkError() {}
