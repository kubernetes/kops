package client

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/spotinst/spotinst-sdk-go/spotinst"
)

// Client provides a client to the API.
type Client struct {
	config *spotinst.Config
}

// New returns a new client.
func New(cfg *spotinst.Config) *Client {
	if cfg == nil {
		cfg = spotinst.DefaultConfig()
	}
	return &Client{cfg}
}

// NewRequest is used to create a new request.
func NewRequest(method, path string) *Request {
	return &Request{
		method: method,
		url: &url.URL{
			Path: path,
		},
		header: make(http.Header),
		Params: make(url.Values),
	}
}

// Do runs a request with our client.
func (c *Client) Do(ctx context.Context, r *Request) (*http.Response, error) {
	req, err := r.toHTTP(ctx, c.config)
	if err != nil {
		return nil, err
	}
	c.logRequest(req)
	resp, err := c.config.HTTPClient.Do(req)
	c.logResponse(resp)
	return resp, err
}

func (c *Client) logf(format string, args ...interface{}) {
	if c.config.Logger != nil {
		c.config.Logger.Printf(format, args...)
	}
}

const logReqMsg = `SPOTINST: Request "%s %s" details:
---[ REQUEST ]---------------------------------------
%s
-----------------------------------------------------`

func (c *Client) logRequest(req *http.Request) {
	if c.config.Logger != nil && req != nil {
		out, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			c.logf(logReqMsg, req.Method, req.URL, string(out))
		}
	}
}

const logRespMsg = `SPOTINST: Response "%s %s" details:
---[ RESPONSE ]----------------------------------------
%s
-------------------------------------------------------`

func (c *Client) logResponse(resp *http.Response) {
	if c.config.Logger != nil && resp != nil {
		out, err := httputil.DumpResponse(resp, true)
		if err == nil {
			c.logf(logRespMsg, resp.Request.Method, resp.Request.URL, string(out))
		}
	}
}
