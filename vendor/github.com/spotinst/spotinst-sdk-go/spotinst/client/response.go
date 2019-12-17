package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Response struct {
	Request struct {
		ID string `json:"id"`
	} `json:"request"`
	Response struct {
		Errors []responseError   `json:"errors"`
		Items  []json.RawMessage `json:"items"`
	} `json:"response"`
}

type responseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field"`
}

type Error struct {
	Response  *http.Response `json:"-"`
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Field     string         `json:"field"`
	RequestID string         `json:"requestId"`
}

func (e Error) Error() string {
	msg := fmt.Sprintf("%v %v: %d (request: %q) %v: %v",
		e.Response.Request.Method, e.Response.Request.URL,
		e.Response.StatusCode, e.RequestID, e.Code, e.Message)

	if e.Field != "" {
		msg = fmt.Sprintf("%s (field: %v)", msg, e.Field)
	}

	return msg
}

type Errors []Error

func (es Errors) Error() string {
	var stack string
	for _, e := range es {
		stack += e.Error() + "\n"
	}
	return stack
}

// DecodeBody is used to JSON decode a body
func DecodeBody(resp *http.Response, out interface{}) error {
	return json.NewDecoder(resp.Body).Decode(out)
}

// RequireOK is used to verify response status code is a successful one (200 OK)
func RequireOK(resp *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, extractError(resp)
	}
	return resp, nil
}

// extractError is used to extract inner/logical errors from the response
func extractError(resp *http.Response) error {
	buf := bytes.NewBuffer(nil)

	// TeeReader returns a Reader that writes to b what it reads from r.Body.
	reader := io.TeeReader(resp.Body, buf)
	defer resp.Body.Close()
	resp.Body = ioutil.NopCloser(buf)

	var out Response
	if err := json.NewDecoder(reader).Decode(&out); err != nil {
		return err
	}

	var errors Errors
	if errs := out.Response.Errors; len(errs) > 0 {
		for _, err := range errs {
			errors = append(errors, Error{
				Response:  resp,
				RequestID: out.Request.ID,
				Code:      err.Code,
				Message:   err.Message,
				Field:     err.Field,
			})
		}
	} else {
		errors = append(errors, Error{
			Response:  resp,
			RequestID: out.Request.ID,
			Code:      strconv.Itoa(resp.StatusCode),
			Message:   http.StatusText(resp.StatusCode),
		})
	}

	return errors
}
