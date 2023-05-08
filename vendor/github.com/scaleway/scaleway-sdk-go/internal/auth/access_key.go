package auth

import "net/http"

type AccessKeyOnly struct {
	// auth config may contain an access key without being authenticated
	AccessKey string
}

// NewNoAuth return an auth with no authentication method
func NewAccessKeyOnly(accessKey string) *AccessKeyOnly {
	return &AccessKeyOnly{accessKey}
}

func (t *AccessKeyOnly) Headers() http.Header {
	return http.Header{}
}

func (t *AccessKeyOnly) AnonymizedHeaders() http.Header {
	return http.Header{}
}
