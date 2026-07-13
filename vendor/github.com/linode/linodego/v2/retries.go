package linodego

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/http2"
)

const (
	RetryAfterHeaderName      = "Retry-After"
	MaintenanceModeHeaderName = "X-Maintenance-Mode"
	DefaultRetryCount         = 1000
)

// RetryConditional is a type alias for a function that determines if a request should be retried based on the response and error.
type RetryConditional func(*http.Response, error) bool

// RetryAfter is a type alias for a function that determines the duration to wait before retrying based on the response.
type RetryAfter func(*http.Response) (time.Duration, error)

// ConfigureRetries configures http.Client to lock until enough time has passed to retry the request as determined by the Retry-After response header.
// If the Retry-After header is not set, we fall back to the value of SetPollDelay.
func ConfigureRetries(c *Client) {
	c.SetRetryAfter(RespectRetryAfter)
	c.SetRetryCount(DefaultRetryCount)
}

func RespectRetryAfter(resp *http.Response) (time.Duration, error) {
	if resp == nil {
		return 0, nil
	}

	retryAfterStr := resp.Header.Get(RetryAfterHeaderName)
	if retryAfterStr == "" {
		return 0, nil
	}

	retryAfter, err := strconv.Atoi(retryAfterStr)
	if err != nil {
		return 0, err
	}

	duration := time.Duration(retryAfter) * time.Second
	log.Printf("[INFO] Respecting Retry-After Header of %d (%s)", retryAfter, duration)

	return duration, nil
}

// Retry conditions

func LinodeBusyRetryCondition(resp *http.Response, _ error) bool {
	if resp == nil {
		return false
	}

	apiError, ok := getAPIError(resp)
	linodeBusy := ok && apiError.Error() == "Linode busy."
	retry := resp.StatusCode == http.StatusBadRequest && linodeBusy

	return retry
}

func TooManyRequestsRetryCondition(resp *http.Response, _ error) bool {
	if resp == nil {
		return false
	}

	return resp.StatusCode == http.StatusTooManyRequests
}

func ServiceUnavailableRetryCondition(resp *http.Response, _ error) bool {
	if resp == nil {
		return false
	}

	serviceUnavailable := resp.StatusCode == http.StatusServiceUnavailable

	// During maintenance events, the API will return a 503 and add
	// an `X-MAINTENANCE-MODE` header. Don't retry during maintenance
	// events, only for legitimate 503s.
	if serviceUnavailable && resp.Header.Get(MaintenanceModeHeaderName) != "" {
		log.Printf("[INFO] Linode API is under maintenance, request will not be retried - please see status.linode.com for more information")
		return false
	}

	return serviceUnavailable
}

func RequestTimeoutRetryCondition(resp *http.Response, _ error) bool {
	if resp == nil {
		return false
	}

	return resp.StatusCode == http.StatusRequestTimeout
}

func RequestGOAWAYRetryCondition(_ *http.Response, err error) bool {
	return errors.As(err, &http2.GoAwayError{})
}

func RequestNGINXRetryCondition(resp *http.Response, _ error) bool {
	if resp == nil {
		return false
	}

	return resp.StatusCode == http.StatusBadRequest &&
		resp.Header.Get("Server") == "nginx" &&
		resp.Header.Get("Content-Type") == "text/html"
}

// Helper function to extract APIError from response
func getAPIError(resp *http.Response) (*APIError, bool) {
	if resp.Body == nil {
		return nil, false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false
	}

	resp.Body = io.NopCloser(bytes.NewReader(body))

	var apiError APIError

	err = json.Unmarshal(body, &apiError)
	if err != nil {
		return nil, false
	}

	return &apiError, true
}
