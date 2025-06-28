package hcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
)

func getRequest[Schema any](ctx context.Context, client *Client, url string) (Schema, *Response, error) {
	var respBody Schema

	req, err := client.NewRequest(ctx, "GET", url, nil)
	if err != nil {
		return respBody, nil, err
	}

	resp, err := client.Do(req, &respBody)
	if err != nil {
		return respBody, resp, err
	}

	return respBody, resp, nil
}

func postRequest[Schema any](ctx context.Context, client *Client, url string, reqBody any) (Schema, *Response, error) {
	var respBody Schema

	var reqBodyReader io.Reader
	if reqBody != nil {
		reqBodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return respBody, nil, err
		}

		reqBodyReader = bytes.NewReader(reqBodyBytes)
	}

	req, err := client.NewRequest(ctx, "POST", url, reqBodyReader)
	if err != nil {
		return respBody, nil, err
	}

	resp, err := client.Do(req, &respBody)
	if err != nil {
		return respBody, resp, err
	}

	return respBody, resp, nil
}

func putRequest[Schema any](ctx context.Context, client *Client, url string, reqBody any) (Schema, *Response, error) {
	var respBody Schema

	var reqBodyReader io.Reader
	if reqBody != nil {
		reqBodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return respBody, nil, err
		}

		reqBodyReader = bytes.NewReader(reqBodyBytes)
	}

	req, err := client.NewRequest(ctx, "PUT", url, reqBodyReader)
	if err != nil {
		return respBody, nil, err
	}

	resp, err := client.Do(req, &respBody)
	if err != nil {
		return respBody, resp, err
	}

	return respBody, resp, nil
}

func deleteRequest[Schema any](ctx context.Context, client *Client, url string) (Schema, *Response, error) {
	var respBody Schema

	req, err := client.NewRequest(ctx, "DELETE", url, nil)
	if err != nil {
		return respBody, nil, err
	}

	resp, err := client.Do(req, &respBody)
	if err != nil {
		return respBody, resp, err
	}

	return respBody, resp, nil
}

func deleteRequestNoResult(ctx context.Context, client *Client, url string) (*Response, error) {
	req, err := client.NewRequest(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	return client.Do(req, nil)
}
