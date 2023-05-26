package auth

import "net/http"

// Auth implement methods required for authentication.
// Valid authentication are currently a token or no auth.
type Auth interface {
	// Headers returns headers that must be add to the http request
	Headers() http.Header

	// AnonymizedHeaders returns an anonymised version of Headers()
	// This method could be use for logging purpose.
	AnonymizedHeaders() http.Header
}

type headerAnonymizer func(header http.Header) http.Header

var headerAnonymizers = []headerAnonymizer{
	AnonymizeTokenHeaders,
	AnonymizeJWTHeaders,
}

func AnonymizeHeaders(headers http.Header) http.Header {
	for _, anonymizer := range headerAnonymizers {
		headers = anonymizer(headers)
	}

	return headers
}
