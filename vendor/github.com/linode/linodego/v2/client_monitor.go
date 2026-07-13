package linodego

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	// MonitorAPIHost is the default monitor-api host
	MonitorAPIHost = "monitor-api.linode.com"
	// MonitorAPIHostVar is the env var to check for the alternate Monitor API URL
	MonitorAPIHostVar = "MONITOR_API_URL"
	// MonitorAPIVersion is the default API version to use
	MonitorAPIVersion = "v2beta"
	// MonitorAPIVersionVar is the env var to check for the alternate Monitor API version
	MonitorAPIVersionVar = "MONITOR_API_VERSION"
	// MonitorAPIEnvVar is the env var to check for Monitor API token
	MonitorAPIEnvVar = "MONITOR_API_TOKEN"
)

// MonitorClient is a wrapper around the http client
type MonitorClient struct {
	httpClient  *http.Client
	debug       bool
	apiBaseURL  string
	apiProtocol string
	apiVersion  string
	hostURL     string
	userAgent   string
	header      http.Header
	logger      Logger
}

// NewMonitorClient is the entry point for user to create a new MonitorClient
// It utilizes default values and looks for environment variables to initialize a MonitorClient.
func NewMonitorClient(hc *http.Client) (mClient MonitorClient) {
	if hc != nil {
		mClient.httpClient = hc
	} else {
		mClient.httpClient = &http.Client{}
	}

	// Ensure transport is initialized so SetRootCertificate can configure TLS
	if mClient.httpClient.Transport == nil {
		mClient.httpClient.Transport = &http.Transport{}
	}

	mClient.header = make(http.Header)
	mClient.logger = createLogger()

	mClient.SetUserAgent(DefaultUserAgent)

	baseURL, baseURLExists := os.LookupEnv(MonitorAPIHostVar)
	if baseURLExists {
		mClient.SetBaseURL(baseURL)
	} else {
		mClient.SetBaseURL(MonitorAPIHost)
	}

	apiVersion, apiVersionExists := os.LookupEnv(MonitorAPIVersionVar)
	if apiVersionExists {
		mClient.SetAPIVersion(apiVersion)
	} else {
		mClient.SetAPIVersion(MonitorAPIVersion)
	}

	token, apiTokenExists := os.LookupEnv(MonitorAPIEnvVar)
	if apiTokenExists {
		mClient.SetToken(token)
	}

	mClient.SetDebug(envDebug)

	return mClient
}

// SetUserAgent sets a custom user-agent for HTTP requests
func (mc *MonitorClient) SetUserAgent(ua string) *MonitorClient {
	mc.userAgent = ua
	mc.header.Set("User-Agent", ua)

	return mc
}

// SetDebug sets the debug on the client
func (mc *MonitorClient) SetDebug(debug bool) *MonitorClient {
	mc.debug = debug

	return mc
}

// SetLogger allows the user to override the output
// logger for debug logs.
func (mc *MonitorClient) SetLogger(logger Logger) *MonitorClient {
	mc.logger = logger

	return mc
}

// SetBaseURL is the helper function to set base url
func (mc *MonitorClient) SetBaseURL(baseURL string) *MonitorClient {
	baseURLPath, _ := url.Parse(baseURL)

	mc.apiBaseURL = path.Join(baseURLPath.Host, baseURLPath.Path)
	mc.apiProtocol = baseURLPath.Scheme

	mc.updateMonitorHostURL()

	return mc
}

// SetAPIVersion is the helper function to set api version
func (mc *MonitorClient) SetAPIVersion(apiVersion string) *MonitorClient {
	mc.apiVersion = apiVersion

	mc.updateMonitorHostURL()

	return mc
}

// SetRootCertificate adds a root certificate to the underlying TLS client config.
func (mc *MonitorClient) SetRootCertificate(certPath string) error {
	transport, ok := mc.httpClient.Transport.(*http.Transport)
	if !ok {
		err := fmt.Errorf("current transport is not an *http.Transport instance")
		if mc.logger != nil {
			mc.logger.Errorf("%s", err)
		}

		return err
	}

	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	if transport.TLSClientConfig.RootCAs == nil {
		transport.TLSClientConfig.RootCAs = x509.NewCertPool()
	}

	pem, err := os.ReadFile(filepath.Clean(certPath))
	if err != nil {
		if mc.logger != nil {
			mc.logger.Errorf("Failed to read root certificate at %s: %s", certPath, err.Error())
		}

		return fmt.Errorf("failed to read root certificate at %s: %w", certPath, err)
	}

	transport.TLSClientConfig.RootCAs.AppendCertsFromPEM(pem)

	return nil
}

// SetToken sets the API token for all requests from this client
func (mc *MonitorClient) SetToken(token string) *MonitorClient {
	mc.header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return mc
}

// SetHeader sets a custom header to be used in all API requests made with the current client.
// NOTE: Some headers may be overridden by the individual request functions.
func (mc *MonitorClient) SetHeader(name, value string) {
	mc.header.Set(name, value)
}

func (mc *MonitorClient) updateMonitorHostURL() {
	apiProto := APIProto
	baseURL := MonitorAPIHost
	apiVersion := MonitorAPIVersion

	if mc.apiBaseURL != "" {
		baseURL = mc.apiBaseURL
	}

	if mc.apiVersion != "" {
		apiVersion = mc.apiVersion
	}

	if mc.apiProtocol != "" {
		apiProto = mc.apiProtocol
	}

	mc.hostURL = fmt.Sprintf(
		"%s://%s/%s",
		apiProto,
		baseURL,
		url.PathEscape(apiVersion),
	)
}

// doRequest is a generic helper to execute HTTP requests for the MonitorClient
func (mc *MonitorClient) doRequest(ctx context.Context, method, endpoint string, params requestParams) error {
	var bodyReader io.Reader

	if params.Body != nil {
		if _, err := params.Body.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek body: %w", err)
		}

		bodyReader = params.Body
	}

	reqURL := fmt.Sprintf("%s/%s", strings.TrimRight(mc.hostURL, "/"), strings.TrimLeft(endpoint, "/"))

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	for name, values := range mc.header {
		for _, value := range values {
			req.Header.Set(name, value)
		}
	}

	if mc.debug && mc.logger != nil {
		mc.logger.Debugf("Sending request: %s %s", method, reqURL)
	}

	resp, err := mc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	_, err = coupleAPIErrors(resp, nil)
	if err != nil {
		return err
	}

	if mc.debug && mc.logger != nil {
		mc.logger.Debugf("Received response: %s", resp.Status)
	}

	if params.Response != nil {
		if err := json.NewDecoder(resp.Body).Decode(params.Response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
