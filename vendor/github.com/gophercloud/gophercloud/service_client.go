package gophercloud

import (
	"context"
	"io"
	"net/http"
	"strings"
)

// ServiceClient stores details required to interact with a specific service API implemented by a provider.
// Generally, you'll acquire these by calling the appropriate `New` method on a ProviderClient.
type ServiceClient struct {
	// ProviderClient is a reference to the provider that implements this service.
	*ProviderClient

	// Endpoint is the base URL of the service's API, acquired from a service catalog.
	// It MUST end with a /.
	Endpoint string

	// ResourceBase is the base URL shared by the resources within a service's API. It should include
	// the API version and, like Endpoint, MUST end with a / if set. If not set, the Endpoint is used
	// as-is, instead.
	ResourceBase string

	// This is the service client type (e.g. compute, sharev2).
	// NOTE: FOR INTERNAL USE ONLY. DO NOT SET. GOPHERCLOUD WILL SET THIS.
	// It is only exported because it gets set in a different package.
	Type string

	// The microversion of the service to use. Set this to use a particular microversion.
	Microversion string

	// MoreHeaders allows users (or Gophercloud) to set service-wide headers on requests. Put another way,
	// values set in this field will be set on all the HTTP requests the service client sends.
	MoreHeaders map[string]string
}

// ResourceBaseURL returns the base URL of any resources used by this service. It MUST end with a /.
func (client *ServiceClient) ResourceBaseURL() string {
	if client.ResourceBase != "" {
		return client.ResourceBase
	}
	return client.Endpoint
}

// ServiceURL constructs a URL for a resource belonging to this provider.
func (client *ServiceClient) ServiceURL(parts ...string) string {
	return client.ResourceBaseURL() + strings.Join(parts, "/")
}

func (client *ServiceClient) initReqOpts(JSONBody interface{}, JSONResponse interface{}, opts *RequestOpts) {
	if v, ok := (JSONBody).(io.Reader); ok {
		opts.RawBody = v
	} else if JSONBody != nil {
		opts.JSONBody = JSONBody
	}

	if JSONResponse != nil {
		opts.JSONResponse = JSONResponse
	}
}

// GetWithContext calls `Request` with the "GET" HTTP verb.
func (client *ServiceClient) GetWithContext(ctx context.Context, url string, JSONResponse interface{}, opts *RequestOpts) (*http.Response, error) {
	if opts == nil {
		opts = new(RequestOpts)
	}
	client.initReqOpts(nil, JSONResponse, opts)
	return client.RequestWithContext(ctx, "GET", url, opts)
}

// Get is a compatibility wrapper for GetWithContext.
func (client *ServiceClient) Get(url string, JSONResponse interface{}, opts *RequestOpts) (*http.Response, error) {
	return client.GetWithContext(context.Background(), url, JSONResponse, opts)
}

// PostWithContext calls `Request` with the "POST" HTTP verb.
func (client *ServiceClient) PostWithContext(ctx context.Context, url string, JSONBody interface{}, JSONResponse interface{}, opts *RequestOpts) (*http.Response, error) {
	if opts == nil {
		opts = new(RequestOpts)
	}
	client.initReqOpts(JSONBody, JSONResponse, opts)
	return client.RequestWithContext(ctx, "POST", url, opts)
}

// Post is a compatibility wrapper for PostWithContext.
func (client *ServiceClient) Post(url string, JSONBody interface{}, JSONResponse interface{}, opts *RequestOpts) (*http.Response, error) {
	return client.PostWithContext(context.Background(), url, JSONBody, JSONResponse, opts)
}

// PutWithContext calls `Request` with the "PUT" HTTP verb.
func (client *ServiceClient) PutWithContext(ctx context.Context, url string, JSONBody interface{}, JSONResponse interface{}, opts *RequestOpts) (*http.Response, error) {
	if opts == nil {
		opts = new(RequestOpts)
	}
	client.initReqOpts(JSONBody, JSONResponse, opts)
	return client.RequestWithContext(ctx, "PUT", url, opts)
}

// Put is a compatibility wrapper for PurWithContext.
func (client *ServiceClient) Put(url string, JSONBody interface{}, JSONResponse interface{}, opts *RequestOpts) (*http.Response, error) {
	return client.PutWithContext(context.Background(), url, JSONBody, JSONResponse, opts)
}

// PatchWithContext calls `Request` with the "PATCH" HTTP verb.
func (client *ServiceClient) PatchWithContext(ctx context.Context, url string, JSONBody interface{}, JSONResponse interface{}, opts *RequestOpts) (*http.Response, error) {
	if opts == nil {
		opts = new(RequestOpts)
	}
	client.initReqOpts(JSONBody, JSONResponse, opts)
	return client.RequestWithContext(ctx, "PATCH", url, opts)
}

// Patch is a compatibility wrapper for PatchWithContext.
func (client *ServiceClient) Patch(url string, JSONBody interface{}, JSONResponse interface{}, opts *RequestOpts) (*http.Response, error) {
	return client.PatchWithContext(context.Background(), url, JSONBody, JSONResponse, opts)
}

// DeleteWithContext calls `Request` with the "DELETE" HTTP verb.
func (client *ServiceClient) DeleteWithContext(ctx context.Context, url string, opts *RequestOpts) (*http.Response, error) {
	if opts == nil {
		opts = new(RequestOpts)
	}
	client.initReqOpts(nil, nil, opts)
	return client.RequestWithContext(ctx, "DELETE", url, opts)
}

// Delete is a compatibility wrapper for DeleteWithContext.
func (client *ServiceClient) Delete(url string, opts *RequestOpts) (*http.Response, error) {
	return client.DeleteWithContext(context.Background(), url, opts)
}

// HeadWithContext calls `Request` with the "HEAD" HTTP verb.
func (client *ServiceClient) HeadWithContext(ctx context.Context, url string, opts *RequestOpts) (*http.Response, error) {
	if opts == nil {
		opts = new(RequestOpts)
	}
	client.initReqOpts(nil, nil, opts)
	return client.RequestWithContext(ctx, "HEAD", url, opts)
}

// Head is a compatibility wrapper for HeadWithContext.
func (client *ServiceClient) Head(url string, opts *RequestOpts) (*http.Response, error) {
	return client.HeadWithContext(context.Background(), url, opts)
}

func (client *ServiceClient) setMicroversionHeader(opts *RequestOpts) {
	switch client.Type {
	case "compute":
		opts.MoreHeaders["X-OpenStack-Nova-API-Version"] = client.Microversion
	case "sharev2":
		opts.MoreHeaders["X-OpenStack-Manila-API-Version"] = client.Microversion
	case "volume":
		opts.MoreHeaders["X-OpenStack-Volume-API-Version"] = client.Microversion
	case "baremetal":
		opts.MoreHeaders["X-OpenStack-Ironic-API-Version"] = client.Microversion
	case "baremetal-introspection":
		opts.MoreHeaders["X-OpenStack-Ironic-Inspector-API-Version"] = client.Microversion
	}

	if client.Type != "" {
		opts.MoreHeaders["OpenStack-API-Version"] = client.Type + " " + client.Microversion
	}
}

// Request carries out the HTTP operation for the service client
func (client *ServiceClient) RequestWithContext(ctx context.Context, method, url string, options *RequestOpts) (*http.Response, error) {
	if options.MoreHeaders == nil {
		options.MoreHeaders = make(map[string]string)
	}

	if client.Microversion != "" {
		client.setMicroversionHeader(options)
	}

	if len(client.MoreHeaders) > 0 {
		if options == nil {
			options = new(RequestOpts)
		}

		for k, v := range client.MoreHeaders {
			options.MoreHeaders[k] = v
		}
	}
	return client.ProviderClient.RequestWithContext(ctx, method, url, options)
}

// Request is a compatibility wrapper for RequestWithContext.
func (client *ServiceClient) Request(method, url string, options *RequestOpts) (*http.Response, error) {
	return client.RequestWithContext(context.Background(), method, url, options)
}

// ParseResponse is a helper function to parse http.Response to constituents.
func ParseResponse(resp *http.Response, err error) (io.ReadCloser, http.Header, error) {
	if resp != nil {
		return resp.Body, resp.Header, err
	}
	return nil, nil, err
}
