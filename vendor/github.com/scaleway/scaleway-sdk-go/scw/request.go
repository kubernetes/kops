package scw

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/scaleway/scaleway-sdk-go/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/auth"
)

// ScalewayRequest contains all the contents related to performing a request on the Scaleway API.
type ScalewayRequest struct {
	Method  string
	Path    string
	Headers http.Header
	Query   url.Values
	Body    io.Reader

	// request options
	ctx      context.Context
	auth     auth.Auth
	allPages bool
	zones    []Zone
	regions  []Region
}

// getURL constructs a URL based on the base url and the client.
func (req *ScalewayRequest) getURL(baseURL string) (*url.URL, error) {
	url, err := url.Parse(baseURL + req.Path)
	if err != nil {
		return nil, errors.New("invalid url %s: %s", baseURL+req.Path, err)
	}
	url.RawQuery = req.Query.Encode()

	return url, nil
}

// SetBody json marshal the given body and write the json content type
// to the request. It also catches when body is a file.
func (req *ScalewayRequest) SetBody(body interface{}) error {
	var contentType string
	var content io.Reader

	switch b := body.(type) {
	case *File:
		contentType = b.ContentType
		content = b.Content
	case io.Reader:
		contentType = "text/plain"
		content = b
	default:
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		contentType = "application/json"
		content = bytes.NewReader(buf)
	}

	if req.Headers == nil {
		req.Headers = http.Header{}
	}

	req.Headers.Set("Content-Type", contentType)
	req.Body = content

	return nil
}

func (req *ScalewayRequest) apply(opts []RequestOption) {
	for _, opt := range opts {
		opt(req)
	}
}

func (req *ScalewayRequest) validate() error {
	// nothing so far
	return nil
}

func (req *ScalewayRequest) clone() *ScalewayRequest {
	clonedReq := &ScalewayRequest{
		Method:   req.Method,
		Path:     req.Path,
		Headers:  req.Headers.Clone(),
		ctx:      req.ctx,
		auth:     req.auth,
		allPages: req.allPages,
		zones:    req.zones,
	}
	if req.Query != nil {
		clonedReq.Query = url.Values(http.Header(req.Query).Clone())
	}
	return clonedReq
}
