package common

import (
	"net/http"
)

type loggingHandler struct {
	next http.Handler
}

func (l *loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Log.Debugf("[http] %s %s", r.Method, r.URL.RequestURI())
	l.next.ServeHTTP(w, r)
}

func LoggingHTTPHandler(h http.Handler) http.Handler {
	return &loggingHandler{next: h}
}
