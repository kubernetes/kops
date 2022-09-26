package scw

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"math"
	"net"
	"net/http"
	"net/http/httputil"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/auth"
	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/logger"
)

// Client is the Scaleway client which performs API requests.
//
// This client should be passed in the `NewApi` functions whenever an API instance is created.
// Creating a Client is done with the `NewClient` function.
type Client struct {
	httpClient            httpClient
	auth                  auth.Auth
	apiURL                string
	userAgent             string
	defaultOrganizationID *string
	defaultProjectID      *string
	defaultRegion         *Region
	defaultZone           *Zone
	defaultPageSize       *uint32
}

func defaultOptions() []ClientOption {
	return []ClientOption{
		WithoutAuth(),
		WithAPIURL("https://api.scaleway.com"),
		withDefaultUserAgent(userAgent),
	}
}

// NewClient instantiate a new Client object.
//
// Zero or more ClientOption object can be passed as a parameter.
// These options will then be applied to the client.
func NewClient(opts ...ClientOption) (*Client, error) {
	s := newSettings()

	// apply options
	s.apply(append(defaultOptions(), opts...))

	// validate settings
	err := s.validate()
	if err != nil {
		return nil, err
	}

	// dial the API
	if s.httpClient == nil {
		s.httpClient = newHTTPClient()
	}

	// insecure mode
	if s.insecure {
		logger.Debugf("client: using insecure mode")
		setInsecureMode(s.httpClient)
	}

	logger.Debugf("client: using sdk version " + version)

	return &Client{
		auth:                  s.token,
		httpClient:            s.httpClient,
		apiURL:                s.apiURL,
		userAgent:             s.userAgent,
		defaultOrganizationID: s.defaultOrganizationID,
		defaultProjectID:      s.defaultProjectID,
		defaultRegion:         s.defaultRegion,
		defaultZone:           s.defaultZone,
		defaultPageSize:       s.defaultPageSize,
	}, nil
}

// GetDefaultOrganizationID returns the default organization ID
// of the client. This value can be set in the client option
// WithDefaultOrganizationID(). Be aware this value can be empty.
func (c *Client) GetDefaultOrganizationID() (organizationID string, exists bool) {
	if c.defaultOrganizationID != nil {
		return *c.defaultOrganizationID, true
	}
	return "", false
}

// GetDefaultProjectID returns the default project ID
// of the client. This value can be set in the client option
// WithDefaultProjectID(). Be aware this value can be empty.
func (c *Client) GetDefaultProjectID() (projectID string, exists bool) {
	if c.defaultProjectID != nil {
		return *c.defaultProjectID, true
	}
	return "", false
}

// GetDefaultRegion returns the default region of the client.
// This value can be set in the client option
// WithDefaultRegion(). Be aware this value can be empty.
func (c *Client) GetDefaultRegion() (region Region, exists bool) {
	if c.defaultRegion != nil {
		return *c.defaultRegion, true
	}
	return Region(""), false
}

// GetDefaultZone returns the default zone of the client.
// This value can be set in the client option
// WithDefaultZone(). Be aware this value can be empty.
func (c *Client) GetDefaultZone() (zone Zone, exists bool) {
	if c.defaultZone != nil {
		return *c.defaultZone, true
	}
	return Zone(""), false
}

func (c *Client) GetSecretKey() (secretKey string, exists bool) {
	if token, isToken := c.auth.(*auth.Token); isToken {
		return token.SecretKey, isToken
	}
	return "", false
}

func (c *Client) GetAccessKey() (accessKey string, exists bool) {
	if token, isToken := c.auth.(*auth.Token); isToken {
		return token.AccessKey, isToken
	}
	return "", false
}

// GetDefaultPageSize returns the default page size of the client.
// This value can be set in the client option
// WithDefaultPageSize(). Be aware this value can be empty.
func (c *Client) GetDefaultPageSize() (pageSize uint32, exists bool) {
	if c.defaultPageSize != nil {
		return *c.defaultPageSize, true
	}
	return 0, false
}

// Do performs HTTP request(s) based on the ScalewayRequest object.
// RequestOptions are applied prior to doing the request.
func (c *Client) Do(req *ScalewayRequest, res interface{}, opts ...RequestOption) (err error) {
	// apply request options
	req.apply(opts)

	// validate request options
	err = req.validate()
	if err != nil {
		return err
	}

	if req.auth == nil {
		req.auth = c.auth
	}

	if req.allPages {
		return c.doListAll(req, res)
	}

	return c.do(req, res)
}

// requestNumber auto increments on each do().
// This allows easy distinguishing of concurrently performed requests in log.
var requestNumber uint32

// do performs a single HTTP request based on the ScalewayRequest object.
func (c *Client) do(req *ScalewayRequest, res interface{}) (sdkErr error) {
	currentRequestNumber := atomic.AddUint32(&requestNumber, 1)

	if req == nil {
		return errors.New("request must be non-nil")
	}

	// build url
	url, sdkErr := req.getURL(c.apiURL)
	if sdkErr != nil {
		return sdkErr
	}
	logger.Debugf("creating %s request on %s", req.Method, url.String())

	// build request
	httpRequest, err := http.NewRequest(req.Method, url.String(), req.Body)
	if err != nil {
		return errors.Wrap(err, "could not create request")
	}

	httpRequest.Header = req.getAllHeaders(req.auth, c.userAgent, false)

	if req.ctx != nil {
		httpRequest = httpRequest.WithContext(req.ctx)
	}

	if logger.ShouldLog(logger.LogLevelDebug) {
		// Keep original headers (before anonymization)
		originalHeaders := httpRequest.Header

		// Get anonymized headers
		httpRequest.Header = req.getAllHeaders(req.auth, c.userAgent, true)

		dump, err := httputil.DumpRequestOut(httpRequest, true)
		if err != nil {
			logger.Warningf("cannot dump outgoing request: %s", err)
		} else {
			var logString string
			logString += "\n--------------- Scaleway SDK REQUEST %d : ---------------\n"
			logString += "%s\n"
			logString += "---------------------------------------------------------"

			logger.Debugf(logString, currentRequestNumber, dump)
		}

		// Restore original headers before sending the request
		httpRequest.Header = originalHeaders
	}

	// execute request
	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return errors.Wrap(err, "error executing request")
	}

	defer func() {
		closeErr := httpResponse.Body.Close()
		if sdkErr == nil && closeErr != nil {
			sdkErr = errors.Wrap(closeErr, "could not close http response")
		}
	}()
	if logger.ShouldLog(logger.LogLevelDebug) {
		dump, err := httputil.DumpResponse(httpResponse, true)
		if err != nil {
			logger.Warningf("cannot dump ingoing response: %s", err)
		} else {
			var logString string
			logString += "\n--------------- Scaleway SDK RESPONSE %d : ---------------\n"
			logString += "%s\n"
			logString += "----------------------------------------------------------"

			logger.Debugf(logString, currentRequestNumber, dump)
		}
	}

	sdkErr = hasResponseError(httpResponse)
	if sdkErr != nil {
		return sdkErr
	}

	if res != nil {
		contentType := httpResponse.Header.Get("Content-Type")

		switch contentType {
		case "application/json":
			err = json.NewDecoder(httpResponse.Body).Decode(&res)
			if err != nil {
				return errors.Wrap(err, "could not parse %s response body", contentType)
			}
		default:
			buffer, isBuffer := res.(io.Writer)
			if !isBuffer {
				return errors.Wrap(err, "could not handle %s response body with %T result type", contentType, buffer)
			}

			_, err := io.Copy(buffer, httpResponse.Body)
			if err != nil {
				return errors.Wrap(err, "could not copy %s response body", contentType)
			}
		}

		// Handle instance API X-Total-Count header
		xTotalCountStr := httpResponse.Header.Get("X-Total-Count")
		if legacyLister, isLegacyLister := res.(legacyLister); isLegacyLister && xTotalCountStr != "" {
			xTotalCount, err := strconv.Atoi(xTotalCountStr)
			if err != nil {
				return errors.Wrap(err, "could not parse X-Total-Count header")
			}
			legacyLister.UnsafeSetTotalCount(xTotalCount)
		}
	}

	return nil
}

type lister interface {
	UnsafeGetTotalCount() uint32
	UnsafeAppend(interface{}) (uint32, error)
}

type legacyLister interface {
	UnsafeSetTotalCount(totalCount int)
}

const maxPageCount uint32 = math.MaxUint32

// doListAll collects all pages of a List request and aggregate all results on a single response.
func (c *Client) doListAll(req *ScalewayRequest, res interface{}) (err error) {
	// check for lister interface
	if response, isLister := res.(lister); isLister {
		pageCount := maxPageCount
		for page := uint32(1); page <= pageCount; page++ {
			// set current page
			req.Query.Set("page", strconv.FormatUint(uint64(page), 10))

			// request the next page
			nextPage := newVariableFromType(response)
			err := c.do(req, nextPage)
			if err != nil {
				return err
			}

			// append results
			pageSize, err := response.UnsafeAppend(nextPage)
			if err != nil {
				return err
			}

			if pageSize == 0 {
				return nil
			}

			// set total count on first request
			if pageCount == maxPageCount {
				totalCount := nextPage.(lister).UnsafeGetTotalCount()
				pageCount = (totalCount + pageSize - 1) / pageSize
			}
		}
		return nil
	}

	return errors.New("%T does not support pagination", res)
}

// newVariableFromType returns a variable set to the zero value of the given type
func newVariableFromType(t interface{}) interface{} {
	// reflect.New always create a pointer, that's why we use reflect.Indirect before
	return reflect.New(reflect.Indirect(reflect.ValueOf(t)).Type()).Interface()
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			MaxIdleConnsPerHost:   20,
		},
	}
}

func setInsecureMode(c httpClient) {
	standardHTTPClient, ok := c.(*http.Client)
	if !ok {
		logger.Warningf("client: cannot use insecure mode with HTTP client of type %T", c)
		return
	}
	transportClient, ok := standardHTTPClient.Transport.(*http.Transport)
	if !ok {
		logger.Warningf("client: cannot use insecure mode with Transport client of type %T", standardHTTPClient.Transport)
		return
	}
	if transportClient.TLSClientConfig == nil {
		transportClient.TLSClientConfig = &tls.Config{}
	}
	transportClient.TLSClientConfig.InsecureSkipVerify = true
}
