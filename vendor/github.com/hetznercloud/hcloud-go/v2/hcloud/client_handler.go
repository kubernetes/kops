package hcloud

import (
	"context"
	"net/http"
)

// handler is an interface representing a client request transaction. The handler are
// meant to be chained, similarly to the [http.RoundTripper] interface.
//
// The handler chain is placed between the [Client] API operations and the
// [http.Client].
type handler interface {
	Do(req *http.Request, v any) (resp *Response, err error)
}

// assembleHandlerChain assembles the chain of handlers used to make API requests.
//
// The order of the handlers is important.
func assembleHandlerChain(client *Client) handler {
	// Start down the chain: sending the http request
	h := newHTTPHandler(client.httpClient)

	// Insert debug writer if enabled
	if client.debugWriter != nil {
		h = wrapDebugHandler(h, client.debugWriter)
	}

	// Read rate limit headers
	h = wrapRateLimitHandler(h)

	// Build error from response
	h = wrapErrorHandler(h)

	// Retry request if condition are met
	h = wrapRetryHandler(h, client.retryBackoffFunc, client.retryMaxRetries)

	// Finally parse the response body into the provided schema
	h = wrapParseHandler(h)

	return h
}

// cloneRequest clones both the request and the request body.
func cloneRequest(req *http.Request, ctx context.Context) (cloned *http.Request, err error) { //revive:disable:context-as-argument
	cloned = req.Clone(ctx)

	if req.ContentLength > 0 {
		cloned.Body, err = req.GetBody()
		if err != nil {
			return nil, err
		}
	}

	return cloned, nil
}
