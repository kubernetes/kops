package auth

import "net/http"

type NoAuth struct {
}

// NewNoAuth return an auth with no authentication method
func NewNoAuth() *NoAuth {
	return &NoAuth{}
}

func (t *NoAuth) Headers() http.Header {
	return http.Header{}
}

func (t *NoAuth) AnonymizedHeaders() http.Header {
	return http.Header{}
}
