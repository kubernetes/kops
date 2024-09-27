package scw

import (
	"math/rand"
	"net/http"
	"net/http/httputil"
	"sync/atomic"

	"github.com/scaleway/scaleway-sdk-go/internal/auth"
	"github.com/scaleway/scaleway-sdk-go/logger"
)

type requestLoggerTransport struct {
	rt http.RoundTripper
	// requestNumber auto increments on each do().
	// This allows easy distinguishing of concurrently performed requests in log.
	requestNumber uint32
}

func (l *requestLoggerTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	currentRequestNumber := atomic.AddUint32(&l.requestNumber, 1)
	// Keep original headers (before anonymization)
	originalHeaders := request.Header

	// Get anonymized headers
	request.Header = auth.AnonymizeHeaders(request.Header.Clone())

	// Add a pseudo random request identifier, this can be used to identify request and response in large logs
	requestIdentifier := rand.Uint32()

	dump, err := httputil.DumpRequestOut(request, true)
	if err != nil {
		logger.Warningf("cannot dump outgoing request: %s", err)
	} else {
		var logString string
		logString += "\n---------- Scaleway SDK REQUEST %d (%x) : ----------\n"
		logString += "%s\n"
		logString += "---------------------------------------------------------\n"

		logger.Debugf(logString, currentRequestNumber, requestIdentifier, dump)
	}

	// Restore original headers before sending the request
	request.Header = originalHeaders
	response, requestError := l.rt.RoundTrip(request)
	if requestError != nil {
		_, isSdkError := requestError.(SdkError)
		if !isSdkError {
			return response, requestError
		}
	}

	dump, err = httputil.DumpResponse(response, true)
	if err != nil {
		logger.Warningf("cannot dump ingoing response: %s", err)
	} else {
		var logString string
		logString += "\n---------- Scaleway SDK RESPONSE %d (%x) : ----------\n"
		logString += "%s\n"
		logString += "----------------------------------------------------------\n"

		logger.Debugf(logString, currentRequestNumber, requestIdentifier, dump)
	}

	return response, requestError
}
