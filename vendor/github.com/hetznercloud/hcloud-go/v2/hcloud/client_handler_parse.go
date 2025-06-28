package hcloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

func wrapParseHandler(wrapped handler) handler {
	return &parseHandler{wrapped}
}

type parseHandler struct {
	handler handler
}

func (h *parseHandler) Do(req *http.Request, v any) (resp *Response, err error) {
	// respBody is not needed down the handler chain
	resp, err = h.handler.Do(req, nil)
	if err != nil {
		return resp, err
	}

	if resp.hasJSONBody() {
		// Parse the response meta
		var s schema.MetaResponse
		if err := json.Unmarshal(resp.body, &s); err != nil {
			return resp, fmt.Errorf("hcloud: error reading response meta data: %w", err)
		}
		if s.Meta.Pagination != nil {
			p := PaginationFromSchema(*s.Meta.Pagination)
			resp.Meta.Pagination = &p
		}
	}

	// Parse the response schema
	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, bytes.NewReader(resp.body))
		} else {
			err = json.Unmarshal(resp.body, v)
		}
	}

	return resp, err
}
