package hcloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

var ErrStatusCode = errors.New("server responded with status code")

func wrapErrorHandler(wrapped handler) handler {
	return &errorHandler{wrapped}
}

type errorHandler struct {
	handler handler
}

func (h *errorHandler) Do(req *http.Request, v any) (resp *Response, err error) {
	resp, err = h.handler.Do(req, v)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode >= 400 && resp.StatusCode <= 599 {
		err = errorFromBody(resp)
		if err == nil {
			err = fmt.Errorf("hcloud: %w %d", ErrStatusCode, resp.StatusCode)
		}
	}
	return resp, err
}

func errorFromBody(resp *Response) error {
	if !resp.hasJSONBody() {
		return nil
	}

	var s schema.ErrorResponse
	if err := json.Unmarshal(resp.body, &s); err != nil {
		return nil // nolint: nilerr
	}
	if s.Error.Code == "" && s.Error.Message == "" {
		return nil
	}

	hcErr := ErrorFromSchema(s.Error)
	hcErr.response = resp
	return hcErr
}
