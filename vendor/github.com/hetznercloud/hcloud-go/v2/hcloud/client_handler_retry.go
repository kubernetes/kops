package hcloud

import (
	"errors"
	"net"
	"net/http"
	"time"
)

func wrapRetryHandler(wrapped handler, backoffFunc BackoffFunc, maxRetries int) handler {
	return &retryHandler{wrapped, backoffFunc, maxRetries}
}

type retryHandler struct {
	handler     handler
	backoffFunc BackoffFunc
	maxRetries  int
}

func (h *retryHandler) Do(req *http.Request, v any) (resp *Response, err error) {
	retries := 0
	ctx := req.Context()

	for {
		// Clone the request using the original context
		cloned, err := cloneRequest(req, ctx)
		if err != nil {
			return nil, err
		}

		resp, err = h.handler.Do(cloned, v)
		if err != nil {
			// Beware the diversity of the errors:
			// - request preparation
			// - network connectivity
			// - http status code (see [errorHandler])
			if ctx.Err() != nil {
				// early return if the context was canceled or timed out
				return resp, err
			}

			if retries < h.maxRetries && retryPolicy(resp, err) {
				select {
				case <-ctx.Done():
					return resp, err
				case <-time.After(h.backoffFunc(retries)):
					retries++
					continue
				}
			}
		}

		return resp, err
	}
}

func retryPolicy(resp *Response, err error) bool {
	if err != nil {
		var apiErr Error
		var netErr net.Error

		switch {
		case errors.As(err, &apiErr):
			switch apiErr.Code { //nolint:exhaustive
			case ErrorCodeConflict:
				return true
			case ErrorCodeRateLimitExceeded:
				return true
			}
		case errors.Is(err, ErrStatusCode):
			switch resp.Response.StatusCode {
			// 5xx errors
			case http.StatusBadGateway, http.StatusGatewayTimeout:
				return true
			}
		case errors.As(err, &netErr):
			if netErr.Timeout() {
				return true
			}
		}
	}

	return false
}
