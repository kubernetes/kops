package api

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	WeaveHTTPHost = "127.0.0.1"
	WeaveHTTPPort = 6784
)

type Client struct {
	baseURL string
	log     Logger
}

func (client *Client) httpVerb(verb string, url string, values url.Values) (string, error) {
	url = client.baseURL + url
	client.log.Debugf("weave %s to %s with %v", verb, url, values)
	var body io.Reader
	if values != nil {
		body = strings.NewReader(values.Encode())
	}
	req, err := http.NewRequest(verb, url, body)
	if err != nil {
		return "", err
	}
	if values != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return string(rbody), nil
	}
	return "", errors.New(resp.Status + ": " + string(rbody))
}

func NewClient(addr string, log Logger) *Client {
	host := WeaveHTTPHost
	port := fmt.Sprintf("%d", WeaveHTTPPort)
	switch parts := strings.Split(addr, ":"); len(parts) {
	case 0:
	case 1:
		if parts[0] != "" {
			host = parts[0]
		}
	case 2:
		if parts[0] != "" {
			host = parts[0]
		}
		if parts[1] != "" {
			port = parts[1]
		}
	default:
		return &Client{baseURL: fmt.Sprintf("http://%s", addr), log: log}
	}
	return &Client{baseURL: fmt.Sprintf("http://%s:%s", host, port), log: log}
}

func (client *Client) Connect(remote string) error {
	_, err := client.httpVerb("POST", "/connect", url.Values{"peer": {remote}})
	return err
}

// IsReady returns true if the API server is up and running
func (client *Client) IsReady() bool {
	_, err := client.httpVerb("GET", "/status", nil)

	return err == nil
}

// WaitAPIServer waits until the API server is ready to serve.
func (client *Client) WaitAPIServer(n int) {
	for i := 0; i < n && !client.IsReady(); i++ {
		time.Sleep(2 * time.Second)
	}
}

type Logger interface {
	Infof(string, ...interface{})
	Debugf(string, ...interface{})
}
