package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/spotinst/spotinst-sdk-go/spotinst"
)

type Request struct {
	Obj    interface{}
	Params url.Values
	url    *url.URL
	method string
	body   io.Reader
	header http.Header
}

// toHTTP converts the request to an HTTP request.
func (r *Request) toHTTP(ctx context.Context, cfg *spotinst.Config) (*http.Request, error) {
	// Set the user credentials.
	creds, err := cfg.Credentials.Get()
	if err != nil {
		return nil, err
	}
	if creds.Token != "" {
		r.header.Set("Authorization", "Bearer "+creds.Token)
	}
	if creds.Account != "" {
		r.Params.Set("accountId", creds.Account)
	}

	// Encode the query parameters.
	r.url.RawQuery = r.Params.Encode()

	// Check if we should encode the body.
	if r.body == nil && r.Obj != nil {
		if b, err := EncodeBody(r.Obj); err != nil {
			return nil, err
		} else {
			r.body = b
		}
	}

	// Create the HTTP request.
	req, err := http.NewRequest(r.method, r.url.RequestURI(), r.body)
	if err != nil {
		return nil, err
	}

	// Set request base URL.
	req.URL.Host = cfg.BaseURL.Host
	req.URL.Scheme = cfg.BaseURL.Scheme

	// Set request headers.
	req.Host = cfg.BaseURL.Host
	req.Header = r.header
	req.Header.Set("Content-Type", cfg.ContentType)
	req.Header.Add("Accept", cfg.ContentType)
	req.Header.Add("User-Agent", cfg.UserAgent)

	return req.WithContext(ctx), nil
}

// EncodeBody is used to encode a request body
func EncodeBody(obj interface{}) (io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(obj); err != nil {
		return nil, err
	}
	return buf, nil
}
