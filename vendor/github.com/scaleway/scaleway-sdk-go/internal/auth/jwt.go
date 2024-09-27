package auth

import (
	"net/http"
	"strings"
)

// JWT is the session token used in browser.
type JWT struct {
	Token string
}

// XSessionTokenHeader is Scaleway auth header for browser
const XSessionTokenHeader = "X-Session-Token" // #nosec G101

// NewJWT create a token authentication from a jwt
func NewJWT(token string) *JWT {
	return &JWT{Token: token}
}

// Headers returns headers that must be added to the http request
func (j *JWT) Headers() http.Header {
	headers := http.Header{}
	headers.Set(XSessionTokenHeader, j.Token)
	return headers
}

func AnonymizeJWTHeaders(headers http.Header) http.Header {
	token := headers.Get(XSessionTokenHeader)

	if token != "" {
		headers.Set(XSessionTokenHeader, HideJWT(token))
	}

	return headers
}

// AnonymizedHeaders returns an anonymized version of Headers()
// This method could be used for logging purpose.
func (j *JWT) AnonymizedHeaders() http.Header {
	return AnonymizeJWTHeaders(j.Headers())
}

func HideJWT(token string) string {
	if len(token) == 0 {
		return ""
	}
	// token should be (header).(payload).(signature)
	lastDot := strings.LastIndex(token, ".")
	if lastDot != -1 {
		token = token[:lastDot]
	}

	return token
}
