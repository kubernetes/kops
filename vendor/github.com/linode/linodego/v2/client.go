package linodego

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"
)

const (
	// APIConfigEnvVar environment var to get path to Linode config
	APIConfigEnvVar = "LINODE_CONFIG"
	// APIConfigProfileEnvVar specifies the profile to use when loading from a Linode config
	APIConfigProfileEnvVar = "LINODE_PROFILE"
	// APIHost Linode API hostname
	APIHost = "api.linode.com"
	// APIHostVar environment var to check for alternate API URL
	APIHostVar = "LINODE_URL"
	// APIHostCert environment var containing path to CA cert to validate against.
	// Note that the custom CA cannot be configured together with a custom HTTP Transport.
	APIHostCert = "LINODE_CA"
	// APIVersion Linode API version
	APIVersion = "v4"
	// APIVersionVar environment var to check for alternate API Version
	APIVersionVar = "LINODE_API_VERSION"
	// APIProto connect to API with http(s)
	APIProto = "https"
	// APIEnvVar environment var to check for API token
	APIEnvVar = "LINODE_TOKEN"
	// APISecondsPerPoll how frequently to poll for new Events or Status in WaitFor functions
	APISecondsPerPoll = 3
	// APIRetryMaxWaitTime is the maximum wait time for retries
	APIRetryMaxWaitTime       = time.Duration(30) * time.Second
	APIDefaultCacheExpiration = time.Minute * 15
)

// Embed the log template files
//
//go:embed request_log_template.tmpl
var requestTemplateStr string

//go:embed response_log_template.tmpl
var responseTemplateStr string

var (
	reqLogTemplate  = template.Must(template.New("request").Parse(requestTemplateStr))
	respLogTemplate = template.Must(template.New("response").Parse(responseTemplateStr))
)

type RequestLog struct {
	Request string
	Host    string
	Headers http.Header
	Body    string
}

type ResponseLog struct {
	Status       string
	Proto        string
	ReceivedAt   string
	TimeDuration string
	Headers      http.Header
	Body         string
}

var envDebug = false

// redactHeadersMap is a map of headers that should be redacted in logs,
// mapping the header name to its redacted value.
var redactHeadersMap = map[string]string{
	"Authorization": "Bearer *******************************",
}

// Client is a wrapper around the http client
type Client struct {
	httpClient *http.Client
	userAgent  string
	debug      bool

	pollInterval time.Duration

	baseURL         string
	apiVersion      string
	apiProto        string
	hostURL         string
	header          http.Header
	selectedProfile string
	loadedProfile   string

	configProfiles map[string]ConfigProfile

	// Fields for caching endpoint responses
	shouldCache     bool
	cacheExpiration time.Duration
	cachedEntries   map[string]clientCacheEntry
	cachedEntryLock *sync.RWMutex
	logger          Logger
	requestLog      func(*RequestLog) error
	onBeforeRequest []func(*http.Request) error
	onAfterResponse []func(*http.Response) error

	retryConditionals []RetryConditional
	retryMaxWaitTime  time.Duration
	retryMinWaitTime  time.Duration
	retryAfter        RetryAfter
	retryCount        int
}

type EnvDefaults struct {
	Token   string
	Profile string
}

type clientCacheEntry struct {
	Created time.Time
	Data    any
	// If != nil, use this instead of the
	// global expiry
	ExpiryOverride *time.Duration
}

type (
	Request  = http.Request
	Response = http.Response
)

func init() {
	if apiDebug, ok := os.LookupEnv("LINODE_DEBUG"); ok {
		if parsed, err := strconv.ParseBool(apiDebug); err == nil {
			envDebug = parsed
			log.Println("[INFO] LINODE_DEBUG being set to", envDebug)
		} else {
			log.Println("[WARN] LINODE_DEBUG should be an integer, 0 or 1")
		}
	}
}

// NewClient factory to create new Client struct.
// nolint:funlen
func NewClient(hc *http.Client) (client Client, err error) {
	if hc != nil {
		client.httpClient = hc
	} else {
		client.httpClient = &http.Client{}
	}

	// Ensure that the Header map is not nil
	if client.httpClient.Transport == nil {
		client.httpClient.Transport = &http.Transport{}
	}

	client.shouldCache = true
	client.cacheExpiration = APIDefaultCacheExpiration
	client.cachedEntries = make(map[string]clientCacheEntry)
	client.cachedEntryLock = &sync.RWMutex{}
	client.configProfiles = make(map[string]ConfigProfile)

	const (
		retryMinWaitDuration = 100 * time.Millisecond
		retryMaxWaitDuration = 2 * time.Second
	)

	client.retryMinWaitTime = retryMinWaitDuration
	client.retryMaxWaitTime = retryMaxWaitDuration

	client.SetUserAgent(DefaultUserAgent)
	client.SetLogger(createLogger())

	baseURL, baseURLExists := os.LookupEnv(APIHostVar)
	if baseURLExists {
		client.SetBaseURL(baseURL)
	}

	apiVersion, apiVersionExists := os.LookupEnv(APIVersionVar)
	if apiVersionExists {
		client.SetAPIVersion(apiVersion)
	} else {
		client.SetAPIVersion(APIVersion)
	}

	certPath, certPathExists := os.LookupEnv(APIHostCert)

	if certPathExists { //nolint:nestif
		if _, ok := client.httpClient.Transport.(*http.Transport); ok {
			if err := client.SetRootCertificate(certPath); err != nil {
				return Client{}, err
			}

			if envDebug {
				log.Printf("[DEBUG] Set API root certificate to %s\n", certPath)
			}
		} else {
			log.Println("[WARN] Custom root certificate is not supported with a custom transport")
		}
	}

	client.
		SetRetryWaitTime(APISecondsPerPoll * time.Second).
		SetPollDelay(APISecondsPerPoll * time.Second).
		SetRetries().
		SetDebug(envDebug).
		enableLogSanitization()

	return client, nil
}

// NewClientFromEnv creates a Client and initializes it with values
// from the LINODE_CONFIG file and the LINODE_TOKEN environment variable.
func NewClientFromEnv(hc *http.Client) (*Client, error) {
	client, err := NewClient(hc)
	if err != nil {
		return nil, err
	}

	// Users are expected to chain NewClient(...) and LoadConfig(...) to customize these options
	configPath, err := resolveValidConfigPath()
	if err != nil {
		return nil, err
	}

	// Populate the token from the environment.
	// Tokens should be first priority to maintain backwards compatibility
	if token, ok := os.LookupEnv(APIEnvVar); ok && token != "" {
		client.SetToken(token)
		return &client, nil
	}

	if p, ok := os.LookupEnv(APIConfigEnvVar); ok {
		configPath = p
	} else if !ok && configPath == "" {
		return nil, fmt.Errorf("no linode config file or token found")
	}

	configProfile := DefaultConfigProfile

	if p, ok := os.LookupEnv(APIConfigProfileEnvVar); ok {
		configProfile = p
	}

	client.selectedProfile = configProfile

	// We should only load the config if the config file exists
	if _, statErr := os.Stat(configPath); statErr != nil {
		return nil, fmt.Errorf("error loading config file %s: %w", configPath, statErr)
	}

	err = client.preLoadConfig(configPath)

	return &client, err
}

// SetUserAgent sets a custom user-agent for HTTP requests
func (c *Client) SetUserAgent(ua string) *Client {
	c.userAgent = ua
	c.SetHeader("User-Agent", c.userAgent)

	return c
}

type requestParams struct {
	Body     *bytes.Reader
	Response any
	// Headers are per-request headers that will be applied only to
	// the individual request, not stored on the shared client state.
	Headers http.Header
}

func (c *Client) ErrorAndLogf(format string, args ...any) error {
	if c.debug && c.logger != nil {
		c.logger.Errorf(format, args...)
	}

	return fmt.Errorf(format, args...)
}

// SetRootCertificate adds a root certificate to the underlying TLS client config.
func (c *Client) SetRootCertificate(certPath string) error {
	config, err := c.tlsConfig()
	if err != nil {
		return fmt.Errorf("custom transport is not allowed with a custom root CA: %w", err)
	}

	if config.RootCAs == nil {
		config.RootCAs = x509.NewCertPool()
	}

	pem, err := os.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return fmt.Errorf("failed to read root certificate at %s: %w", certPath, err)
	}

	config.RootCAs.AppendCertsFromPEM(pem)

	return nil
}

// SetToken sets the API token for all requests from this client
// Only necessary if you haven't already provided the http client to NewClient() configured with the token.
func (c *Client) SetToken(token string) *Client {
	c.SetHeader("Authorization", fmt.Sprintf("Bearer %s", token))
	return c
}

// SetRetries adds retry conditions for "Linode Busy." errors and 429s.
func (c *Client) SetRetries() *Client {
	c.
		AddRetryCondition(LinodeBusyRetryCondition).
		AddRetryCondition(TooManyRequestsRetryCondition).
		AddRetryCondition(ServiceUnavailableRetryCondition).
		AddRetryCondition(RequestTimeoutRetryCondition).
		AddRetryCondition(RequestGOAWAYRetryCondition).
		AddRetryCondition(RequestNGINXRetryCondition).
		SetRetryMaxWaitTime(APIRetryMaxWaitTime)
	ConfigureRetries(c)

	return c
}

// AddRetryCondition adds a RetryConditional function to the Client
func (c *Client) AddRetryCondition(retryCondition RetryConditional) *Client {
	c.retryConditionals = append(c.retryConditionals, retryCondition)

	return c
}

func (c *Client) SetDebug(debug bool) *Client {
	c.debug = debug

	return c
}

func (c *Client) SetLogger(logger Logger) *Client {
	c.logger = logger

	return c
}

func (c *Client) OnBeforeRequest(m func(*http.Request) error) {
	c.onBeforeRequest = append(c.onBeforeRequest, m)
}

func (c *Client) OnAfterResponse(m func(*http.Response) error) {
	c.onAfterResponse = append(c.onAfterResponse, m)
}

// UseURL parses the individual components of the given API URL and configures the client
// accordingly. For example, a valid URL.
// For example:
//
//	client.UseURL("https://api.test.linode.com/v4beta")
func (c *Client) UseURL(apiURL string) (*Client, error) {
	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("need both scheme and host in API URL, got %q", apiURL)
	}

	// Create a new URL excluding the path to use as the base URL
	baseURL := &url.URL{
		Host:   parsedURL.Host,
		Scheme: parsedURL.Scheme,
	}

	c.SetBaseURL(baseURL.String())

	versionMatches := regexp.MustCompile(`/v[a-zA-Z0-9]+`).FindAllString(parsedURL.Path, -1)

	// Only set the version if a version is found in the URL, else use the default
	if len(versionMatches) > 0 {
		c.SetAPIVersion(
			strings.Trim(versionMatches[len(versionMatches)-1], "/"),
		)
	}

	return c, nil
}

func (c *Client) SetBaseURL(baseURL string) *Client {
	baseURLPath, _ := url.Parse(baseURL)

	c.baseURL = path.Join(baseURLPath.Host, baseURLPath.Path)
	c.apiProto = baseURLPath.Scheme

	c.updateHostURL()

	return c
}

// SetAPIVersion sets the version of the API to interface with
func (c *Client) SetAPIVersion(apiVersion string) *Client {
	c.apiVersion = apiVersion

	c.updateHostURL()

	return c
}

// InvalidateCache clears all cached responses for all endpoints.
func (c *Client) InvalidateCache() {
	c.cachedEntryLock.Lock()
	defer c.cachedEntryLock.Unlock()

	// GC will handle the old map
	c.cachedEntries = make(map[string]clientCacheEntry)
}

// InvalidateCacheEndpoint invalidates a single cached endpoint.
func (c *Client) InvalidateCacheEndpoint(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("failed to parse URL for caching: %w", err)
	}

	c.cachedEntryLock.Lock()
	defer c.cachedEntryLock.Unlock()

	delete(c.cachedEntries, u.Path)

	return nil
}

// SetGlobalCacheExpiration sets the desired time for any cached response
// to be valid for.
func (c *Client) SetGlobalCacheExpiration(expiryTime time.Duration) {
	c.cacheExpiration = expiryTime
}

// UseCache sets whether response caching should be used
func (c *Client) UseCache(value bool) {
	c.shouldCache = value
}

// SetRetryMaxWaitTime sets the maximum delay before retrying a request.
func (c *Client) SetRetryMaxWaitTime(maxWaitTime time.Duration) *Client {
	c.retryMaxWaitTime = maxWaitTime
	return c
}

// SetRetryWaitTime sets the default (minimum) delay before retrying a request.
func (c *Client) SetRetryWaitTime(minWaitTime time.Duration) *Client {
	c.retryMinWaitTime = minWaitTime
	return c
}

// SetRetryAfter sets the callback function to be invoked with a failed request
// to determine wben it should be retried.
func (c *Client) SetRetryAfter(callback RetryAfter) *Client {
	c.retryAfter = callback
	return c
}

// SetRetryCount sets the number of retries after the initial request before aborting.
// Negative values are treated as 0 (no retries).
func (c *Client) SetRetryCount(count int) *Client {
	if count < 0 {
		count = 0
	}

	c.retryCount = count

	return c
}

// SetPollDelay sets the number of milliseconds to wait between events or status polls.
// Affects all WaitFor* functions and retries.
func (c *Client) SetPollDelay(delay time.Duration) *Client {
	c.pollInterval = delay
	return c
}

// GetPollDelay gets the number of milliseconds to wait between events or status polls.
// Affects all WaitFor* functions and retries.
func (c *Client) GetPollDelay() time.Duration {
	return c.pollInterval
}

// SetHeader sets a custom header to be used in all API requests made with the current
// client.
// NOTE: Some headers may be overridden by the individual request functions.
func (c *Client) SetHeader(name, value string) {
	if c.header == nil {
		c.header = make(http.Header) // Initialize header if nil
	}

	c.header.Set(name, value)
}

func (c *Client) Transport() (*http.Transport, error) {
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		return transport, nil
	}

	return nil, fmt.Errorf("current transport is not an *http.Transport instance")
}

// Generic helper to execute HTTP requests using the net/http package
//
// nolint:funlen, gocognit, nestif
func (c *Client) doRequest(ctx context.Context, method, endpoint string, params requestParams, paginationMutator *func(*http.Request) error) error {
	var (
		req  *http.Request
		resp *http.Response
		err  error
	)

	// retryCount controls the number of retries after the initial attempt
	for range c.retryCount + 1 {
		// createRequest seeks params.Body back to the start, so it's safe to retry.
		req, err = c.createRequest(ctx, method, endpoint, params)
		if err != nil {
			return err
		}

		if paginationMutator != nil {
			if mutErr := (*paginationMutator)(req); mutErr != nil {
				return c.ErrorAndLogf("failed to mutate before request: %v", mutErr.Error())
			}
		}

		if err = c.applyBeforeRequest(req); err != nil {
			return err
		}

		if c.debug && c.logger != nil {
			req = c.logRequest(req)
		}

		processResponse := func(start, end time.Time) error {
			defer func() {
				closeErr := resp.Body.Close()
				if closeErr != nil && err == nil {
					err = closeErr
				}
			}()

			if err = c.checkHTTPError(resp); err != nil {
				return err
			}

			if c.debug && c.logger != nil {
				resp = c.logResponse(resp, start, end)
			}

			if params.Response != nil {
				if err = c.decodeResponseBody(resp, params.Response); err != nil {
					return err
				}
			}

			// Apply after-response mutations
			if err = c.applyAfterResponse(resp); err != nil {
				return err
			}

			return nil
		}

		startTime := time.Now()
		resp, err = c.sendRequest(req)
		endTime := time.Now()

		if err == nil {
			if err = processResponse(startTime, endTime); err == nil {
				return nil
			}
		}

		if !c.shouldRetry(resp, err) {
			break
		}

		retryAfter, retryErr := c.retryAfter(resp)
		if retryErr != nil {
			return retryErr
		}

		// Determine wait time before retrying.
		// If the server provided a Retry-After duration, use it (clamped to bounds).
		// Otherwise, fall back to the configured minimum wait time.
		waitTime := c.retryMinWaitTime

		if retryAfter > 0 {
			waitTime = retryAfter
		}

		// Ensure the wait time is within the defined bounds
		if waitTime < c.retryMinWaitTime {
			waitTime = c.retryMinWaitTime
		} else if waitTime > c.retryMaxWaitTime {
			waitTime = c.retryMaxWaitTime
		}

		// Sleep for the calculated duration before retrying
		time.Sleep(waitTime)
	}

	return err
}

func (c *Client) shouldRetry(resp *http.Response, err error) bool {
	for _, retryConditional := range c.retryConditionals {
		if retryConditional(resp, err) {
			log.Printf("[INFO] Received error %v - Retrying", err)
			return true
		}
	}

	return false
}

func (c *Client) createRequest(ctx context.Context, method, endpoint string, params requestParams) (*http.Request, error) {
	var bodyReader io.Reader

	if params.Body != nil {
		// Reset the body position to the start before using it
		_, err := params.Body.Seek(0, io.SeekStart)
		if err != nil {
			return nil, c.ErrorAndLogf("failed to seek to the start of the body: %v", err.Error())
		}

		bodyReader = params.Body
	}

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s/%s", strings.TrimRight(c.hostURL, "/"),
		strings.TrimLeft(endpoint, "/")), bodyReader)
	if err != nil {
		return nil, c.ErrorAndLogf("failed to create request: %v", err.Error())
	}

	// Set the default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	// Set additional headers added to the client
	for name, values := range c.header {
		for _, value := range values {
			req.Header.Set(name, value)
		}
	}

	// Apply per-request headers (these take priority over client headers)
	for name, values := range params.Headers {
		for _, value := range values {
			req.Header.Set(name, value)
		}
	}

	return req, nil
}

func (c *Client) applyBeforeRequest(req *http.Request) error {
	for _, mutate := range c.onBeforeRequest {
		if err := mutate(req); err != nil {
			return c.ErrorAndLogf("failed to mutate before request: %v", err.Error())
		}
	}

	return nil
}

func (c *Client) applyAfterResponse(resp *http.Response) error {
	for _, mutate := range c.onAfterResponse {
		if err := mutate(resp); err != nil {
			return c.ErrorAndLogf("failed to mutate after response: %v", err.Error())
		}
	}

	return nil
}

func redactHeaders(headers http.Header) http.Header {
	redacted := headers.Clone()

	for header, redactedValue := range redactHeadersMap {
		if headers.Get(header) != "" {
			redacted.Set(header, redactedValue)
		}
	}

	return redacted
}

func (c *Client) logRequest(req *http.Request) *http.Request {
	var reqBody bytes.Buffer
	if req.Body != nil {
		if _, err := io.Copy(&reqBody, req.Body); err != nil {
			c.logger.Errorf("failed to read request body: %v", err)
		}

		req.Body = io.NopCloser(bytes.NewReader(reqBody.Bytes()))
	}

	reqLog := &RequestLog{
		Request: strings.Join([]string{req.Method, req.URL.Path, req.Proto}, " "),
		Host:    req.Host,
		Headers: redactHeaders(req.Header.Clone()),
		Body:    reqBody.String(),
	}

	e := c.requestLog(reqLog)
	if e != nil {
		_ = c.ErrorAndLogf("failed to log request: %v", e.Error())
	}

	sanitizedBody := sanitizeLogValue(reqLog.Body)

	body, jsonErr := formatBody(sanitizedBody)
	if jsonErr != nil {
		if c.debug && c.logger != nil {
			c.logger.Errorf("%v", jsonErr)
		}
	}

	var logBuf bytes.Buffer

	err := reqLogTemplate.Execute(&logBuf, map[string]any{
		"Request": reqLog.Request,
		"Host":    reqLog.Host,
		"Headers": formatHeaders(reqLog.Headers),
		"Body":    body,
	})
	if err == nil {
		c.logger.Debugf(sanitizeLogValue(logBuf.String()))
	}

	return req
}

func formatHeaders(headers map[string][]string) string {
	var builder strings.Builder
	builder.WriteString("\n")

	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		builder.WriteString(fmt.Sprintf("    %s: %s\n", key, strings.Join(headers[key], ", ")))
	}

	return strings.TrimSuffix(builder.String(), "\n")
}

// sanitizeLogValue removes or escapes control characters that could
// enable log injection (e.g., \r, \n) from a string before it is written
// to a log entry. Uses strings.ReplaceAll so static-analysis tools
// (e.g., CodeQL) can recognize the sanitization.
func sanitizeLogValue(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\n")
	s = strings.ReplaceAll(s, "\n", "\\n")

	return s
}

func formatBody(body string) (string, error) {
	body = strings.TrimSpace(body)
	if body == "null" || body == "nil" || body == "" {
		return "", nil
	}

	var jsonData any
	if err := json.Unmarshal([]byte(body), &jsonData); err != nil {
		return "", fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %w", err)
	}

	return "\n" + string(prettyJSON), nil
}

func formatDate(dateStr string) (string, error) {
	parsedTime, err := time.Parse(time.RFC1123, dateStr)
	if err != nil {
		return "", fmt.Errorf("error parsing date: %v", err)
	}

	formattedDate := parsedTime.In(time.Local).Format("2006-01-02T15:04:05-07:00") // nolint:gosmopolitan

	return formattedDate, nil
}

func (c *Client) sendRequest(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req) //#nosec G704 // URL is constructed from client-configured base URL + endpoint
	if err != nil {
		return nil, c.ErrorAndLogf("failed to send request: %w", err)
	}

	return resp, nil
}

func (c *Client) checkHTTPError(resp *http.Response) error {
	_, err := coupleAPIErrors(resp, nil)
	if err != nil {
		_ = c.ErrorAndLogf("received HTTP error: %v", err.Error())
		return err
	}

	return nil
}

func (c *Client) logResponse(resp *http.Response, start, end time.Time) *http.Response {
	var respBody bytes.Buffer
	if _, err := io.Copy(&respBody, resp.Body); err != nil {
		c.logger.Errorf("failed to read response body: %v", err)
	}

	receivedAt, dateErr := formatDate(resp.Header.Get("Date"))
	if dateErr != nil {
		if c.debug && c.logger != nil {
			c.logger.Errorf("failed to format date: %v", dateErr)
		}
	}

	duration := end.Sub(start).String()

	respLog := &ResponseLog{
		Status:       resp.Status,
		Proto:        resp.Proto,
		ReceivedAt:   receivedAt,
		TimeDuration: duration,
		Headers:      resp.Header,
		Body:         respBody.String(),
	}

	body, jsonErr := formatBody(sanitizeLogValue(respLog.Body))
	if jsonErr != nil {
		if c.debug && c.logger != nil {
			c.logger.Errorf("%v", jsonErr)
		}
	}

	var logBuf bytes.Buffer

	err := respLogTemplate.Execute(&logBuf, map[string]any{
		"Status":       respLog.Status,
		"Proto":        respLog.Proto,
		"ReceivedAt":   respLog.ReceivedAt,
		"TimeDuration": respLog.TimeDuration,
		"Headers":      formatHeaders(redactHeaders(respLog.Headers)),
		"Body":         body,
	})
	if err == nil {
		c.logger.Debugf(sanitizeLogValue(logBuf.String()))
	}

	resp.Body = io.NopCloser(bytes.NewReader(respBody.Bytes()))

	return resp
}

func (c *Client) decodeResponseBody(resp *http.Response, response any) error {
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return c.ErrorAndLogf("failed to decode response: %v", err.Error())
	}

	return nil
}

func (c *Client) updateHostURL() {
	apiProto := APIProto
	baseURL := APIHost
	apiVersion := APIVersion

	if c.baseURL != "" {
		baseURL = c.baseURL
	}

	if c.apiVersion != "" {
		apiVersion = c.apiVersion
	}

	if c.apiProto != "" {
		apiProto = c.apiProto
	}

	c.hostURL = strings.TrimRight(fmt.Sprintf("%s://%s/%s", apiProto, baseURL, url.PathEscape(apiVersion)), "/")
}

func (c *Client) tlsConfig() (*tls.Config, error) {
	transport, err := c.Transport()
	if err != nil {
		return nil, err
	}

	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return transport.TLSClientConfig, nil
}

func (c *Client) addCachedResponse(endpoint string, response any, expiry *time.Duration) {
	if !c.shouldCache {
		return
	}

	responseValue := reflect.ValueOf(response)

	entry := clientCacheEntry{
		Created:        time.Now(),
		ExpiryOverride: expiry,
	}

	switch responseValue.Kind() {
	case reflect.Pointer:
		// We want to automatically deref pointers to
		// avoid caching mutable data.
		entry.Data = responseValue.Elem().Interface()
	default:
		entry.Data = response
	}

	c.cachedEntryLock.Lock()
	defer c.cachedEntryLock.Unlock()

	c.cachedEntries[endpoint] = entry
}

func (c *Client) getCachedResponse(endpoint string) any {
	if !c.shouldCache {
		return nil
	}

	c.cachedEntryLock.RLock()

	// Hacky logic to dynamically RUnlock
	// only if it is still locked by the
	// end of the function.
	// This is necessary as we take write
	// access if the entry has expired.
	rLocked := true

	defer func() {
		if rLocked {
			c.cachedEntryLock.RUnlock()
		}
	}()

	entry, ok := c.cachedEntries[endpoint]
	if !ok {
		return nil
	}

	// Handle expired entries
	elapsedTime := time.Since(entry.Created)

	hasExpired := elapsedTime > c.cacheExpiration
	if entry.ExpiryOverride != nil {
		hasExpired = elapsedTime > *entry.ExpiryOverride
	}

	if hasExpired {
		// We need to give up our read access and request read-write access
		c.cachedEntryLock.RUnlock()

		rLocked = false

		c.cachedEntryLock.Lock()
		defer c.cachedEntryLock.Unlock()

		delete(c.cachedEntries, endpoint)

		return nil
	}

	return c.cachedEntries[endpoint].Data
}

func (c *Client) onRequestLog(rl func(*RequestLog) error) *Client {
	if c.requestLog != nil {
		c.logger.Warnf("Overwriting an existing on-request-log callback from=%s to=%s",
			functionName(c.requestLog), functionName(rl))
	}

	c.requestLog = rl

	return c
}

func functionName(i any) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func (c *Client) enableLogSanitization() *Client {
	c.onRequestLog(func(r *RequestLog) error {
		// masking authorization header
		r.Headers.Set("Authorization", "Bearer *******************************")
		return nil
	})

	return c
}

func (c *Client) preLoadConfig(configPath string) error {
	if envDebug {
		log.Printf("[INFO] Loading profile from %s\n", configPath)
	}

	if err := c.LoadConfig(&LoadConfigOptions{
		Path:            configPath,
		SkipLoadProfile: true,
	}); err != nil {
		return err
	}

	// We don't want to load the profile until the user is actually making requests
	c.OnBeforeRequest(func(_ *Request) error {
		if c.loadedProfile != c.selectedProfile {
			if err := c.UseProfile(c.selectedProfile); err != nil {
				return err
			}
		}

		return nil
	})

	return nil
}

func copyBool(bPtr *bool) *bool {
	if bPtr == nil {
		return nil
	}

	t := *bPtr

	return &t
}

func copyInt(iPtr *int) *int {
	if iPtr == nil {
		return nil
	}

	t := *iPtr

	return &t
}

func copyString(sPtr *string) *string {
	if sPtr == nil {
		return nil
	}

	t := *sPtr

	return &t
}

// copyValue returns a pointer to a new value copied from the value
// at the given pointer.
func copyValue[T any](ptr *T) *T {
	if ptr == nil {
		return nil
	}

	t := *ptr

	return &t
}

func copyTime(tPtr *time.Time) *time.Time {
	if tPtr == nil {
		return nil
	}

	t := *tPtr

	return &t
}

func generateListCacheURL(endpoint string, opts *ListOptions) (string, error) {
	if opts == nil {
		return endpoint, nil
	}

	hashedOpts, err := opts.Hash()
	if err != nil {
		return endpoint, err
	}

	return fmt.Sprintf("%s:%s", endpoint, hashedOpts), nil
}
