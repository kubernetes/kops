package hcloud

import (
	"net/http"
)

func newHTTPHandler(httpClient *http.Client) handler {
	return &httpHandler{httpClient}
}

type httpHandler struct {
	httpClient *http.Client
}

func (h *httpHandler) Do(req *http.Request, _ interface{}) (*Response, error) {
	httpResponse, err := h.httpClient.Do(req) //nolint: bodyclose
	resp := &Response{Response: httpResponse}
	if err != nil {
		return resp, err
	}

	err = resp.populateBody()
	if err != nil {
		return resp, err
	}

	return resp, err
}
