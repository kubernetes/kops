package proxy

import "net/http"

type interceptor interface {
	InterceptRequest(*http.Request) error
	InterceptResponse(*http.Response) error
}

type nullInterceptor struct {
}

func (i nullInterceptor) InterceptRequest(r *http.Request) error {
	return nil
}

func (i nullInterceptor) InterceptResponse(r *http.Response) error {
	return nil
}
