package scw

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/scaleway/scaleway-sdk-go/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/auth"
	"github.com/scaleway/scaleway-sdk-go/internal/generic"
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
		logger.Debugf("client: using insecure mode\n")
		setInsecureMode(s.httpClient)
	}

	if logger.ShouldLog(logger.LogLevelDebug) {
		logger.Debugf("client: using request logger\n")
		setRequestLogging(s.httpClient)
	}

	logger.Debugf("client: using sdk version " + getVersion() + "\n")

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
	} else if token, isAccessKey := c.auth.(*auth.AccessKeyOnly); isAccessKey {
		return token.AccessKey, isAccessKey
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
func (c *Client) Do(req *ScalewayRequest, res any, opts ...RequestOption) (err error) {
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

	if req.zones != nil {
		return c.doListZones(req, res, req.zones)
	}
	if req.regions != nil {
		return c.doListRegions(req, res, req.regions)
	}

	if req.allPages {
		return c.doListAll(req, res)
	}

	return c.do(req, res)
}

// do performs a single HTTP request based on the ScalewayRequest object.
func (c *Client) do(req *ScalewayRequest, res any) (sdkErr error) {
	if req == nil {
		return errors.New("request must be non-nil")
	}

	// build url
	url, sdkErr := req.getURL(c.apiURL)
	if sdkErr != nil {
		return sdkErr
	}
	logger.Debugf("creating %s request on %s\n", req.Method, url.String())

	// build request
	ctx := req.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	httpRequest, err := http.NewRequestWithContext(ctx, req.Method, url.String(), req.Body)
	if err != nil {
		return errors.Wrap(err, "could not create request")
	}

	httpRequest.Header = req.getAllHeaders(req.auth, c.userAgent, false)

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

	sdkErr = hasResponseError(httpResponse)
	if sdkErr != nil {
		return sdkErr
	}

	if res != nil && httpResponse.ContentLength != 0 {
		contentType := httpResponse.Header.Get("Content-Type")

		if strings.HasPrefix(contentType, "application/json") {
			err = json.NewDecoder(httpResponse.Body).Decode(&res)
			if err != nil {
				return errors.Wrap(err, "could not parse %s response body", contentType)
			}
		} else {
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
			xTotalCount, err := strconv.ParseInt(xTotalCountStr, 10, 32)
			if err != nil {
				return errors.Wrap(err, "could not parse X-Total-Count header")
			}
			legacyLister.UnsafeSetTotalCount(int(xTotalCount))
		}
	}

	return nil
}

type lister interface {
	UnsafeGetTotalCount() uint64
	UnsafeAppend(any) (uint64, error)
}

// Old lister for uint32
// Used for retro-compatibility with response that use uint32
type lister32 interface {
	UnsafeGetTotalCount() uint32
	UnsafeAppend(any) (uint32, error)
}

type legacyLister interface {
	UnsafeSetTotalCount(totalCount int)
}

func listerGetTotalCount(i any) uint64 {
	if l, isLister := i.(lister); isLister {
		return l.UnsafeGetTotalCount()
	}
	if l32, isLister32 := i.(lister32); isLister32 {
		return uint64(l32.UnsafeGetTotalCount())
	}
	panic(fmt.Errorf("%T does not support pagination but checks failed, should not happen", i))
}

func listerAppend(recv any, elems any) (uint64, error) {
	if l, isLister := recv.(lister); isLister {
		return l.UnsafeAppend(elems)
	} else if l32, isLister32 := recv.(lister32); isLister32 {
		total, err := l32.UnsafeAppend(elems)
		return uint64(total), err
	}

	panic(fmt.Errorf("%T does not support pagination but checks failed, should not happen", recv))
}

func isLister(i any) bool {
	switch i.(type) {
	case lister:
		return true
	case lister32:
		return true
	default:
		return false
	}
}

const maxPageCount uint64 = math.MaxUint32

// doListAll collects all pages of a List request and aggregate all results on a single response.
func (c *Client) doListAll(req *ScalewayRequest, res any) (err error) {
	// check for lister interface
	if isLister(res) {
		pageCount := maxPageCount
		for page := uint64(1); page <= pageCount; page++ {
			// set current page
			req.Query.Set("page", strconv.FormatUint(page, 10))

			// request the next page
			nextPage := newVariableFromType(res)
			err := c.do(req, nextPage)
			if err != nil {
				return err
			}

			// append results
			pageSize, err := listerAppend(res, nextPage)
			if err != nil {
				return err
			}

			if pageSize == 0 {
				return nil
			}

			// set total count on first request
			if pageCount == maxPageCount {
				totalCount := listerGetTotalCount(nextPage)
				pageCount = (totalCount + pageSize - 1) / pageSize
			}
		}
		return nil
	}

	return errors.New("%T does not support pagination", res)
}

// doListLocalities collects all localities using multiple list requests and aggregate all results on a lister response
// results is sorted by locality
func (c *Client) doListLocalities(req *ScalewayRequest, res any, localities []string) (err error) {
	path := req.Path
	if !strings.Contains(path, "%locality%") {
		return errors.New("request is not a valid locality request")
	}
	// Requests are parallelized
	responseMutex := sync.Mutex{}
	requestGroup := sync.WaitGroup{}
	errChan := make(chan error, len(localities))

	requestGroup.Add(len(localities))
	for _, locality := range localities {
		go func(locality string) {
			defer requestGroup.Done()
			// Request is cloned as doListAll will change header
			// We remove zones as it would recurse in the same function
			req := req.clone()
			req.zones = []Zone(nil)
			req.Path = strings.ReplaceAll(path, "%locality%", locality)

			// We create a new response that we append to main response
			zoneResponse := newVariableFromType(res)
			err := c.Do(req, zoneResponse)
			if err != nil {
				errChan <- err
			}
			responseMutex.Lock()
			_, err = listerAppend(res, zoneResponse)
			responseMutex.Unlock()
			if err != nil {
				errChan <- err
			}
		}(locality)
	}
	requestGroup.Wait()

L: // We gather potential errors and return them all together
	for {
		select {
		case newErr := <-errChan:
			err = errors.Wrap(err, "%s", newErr.Error())
		default:
			break L
		}
	}
	close(errChan)
	if err != nil {
		return err
	}
	return nil
}

// doListZones collects all zones using multiple list requests and aggregate all results on a single response.
// result is sorted by zone
func (c *Client) doListZones(req *ScalewayRequest, res any, zones []Zone) (err error) {
	if isLister(res) {
		// Prepare request with %zone% that can be replaced with actual zone
		for _, zone := range AllZones {
			if strings.Contains(req.Path, string(zone)) {
				req.Path = strings.ReplaceAll(req.Path, string(zone), "%locality%")
				break
			}
		}
		if !strings.Contains(req.Path, "%locality%") {
			return errors.New("request is not a valid zoned request")
		}
		localities := make([]string, 0, len(zones))
		for _, zone := range zones {
			localities = append(localities, string(zone))
		}

		err := c.doListLocalities(req, res, localities)
		if err != nil {
			return fmt.Errorf("failed to list localities: %w", err)
		}

		sortResponseByZones(res, zones)
		return nil
	}

	return errors.New("%T does not support pagination", res)
}

// doListRegions collects all regions using multiple list requests and aggregate all results on a single response.
// result is sorted by region
func (c *Client) doListRegions(req *ScalewayRequest, res any, regions []Region) (err error) {
	if isLister(res) {
		// Prepare request with %locality% that can be replaced with actual region
		for _, region := range AllRegions {
			if strings.Contains(req.Path, string(region)) {
				req.Path = strings.ReplaceAll(req.Path, string(region), "%locality%")
				break
			}
		}
		if !strings.Contains(req.Path, "%locality%") {
			return errors.New("request is not a valid zoned request")
		}
		localities := make([]string, 0, len(regions))
		for _, region := range regions {
			localities = append(localities, string(region))
		}

		err := c.doListLocalities(req, res, localities)
		if err != nil {
			return fmt.Errorf("failed to list localities: %w", err)
		}

		sortResponseByRegions(res, regions)
		return nil
	}

	return errors.New("%T does not support pagination", res)
}

// sortSliceByZones sorts a slice of struct using a Zone field that should exist
func sortSliceByZones(list any, zones []Zone) {
	if !generic.HasField(list, "Zone") {
		return
	}

	zoneMap := map[Zone]int{}
	for i, zone := range zones {
		zoneMap[zone] = i
	}
	generic.SortSliceByField(list, "Zone", func(i any, i2 any) bool {
		return zoneMap[i.(Zone)] < zoneMap[i2.(Zone)]
	})
}

// sortSliceByRegions sorts a slice of struct using a Region field that should exist
func sortSliceByRegions(list any, regions []Region) {
	if !generic.HasField(list, "Region") {
		return
	}

	regionMap := map[Region]int{}
	for i, region := range regions {
		regionMap[region] = i
	}
	generic.SortSliceByField(list, "Region", func(i any, i2 any) bool {
		return regionMap[i.(Region)] < regionMap[i2.(Region)]
	})
}

// sortResponseByZones find first field that is a slice in a struct and sort it by zone
func sortResponseByZones(res any, zones []Zone) {
	// res may be ListServersResponse
	//
	// type ListServersResponse struct {
	//	TotalCount uint32 `json:"total_count"`
	//	Servers []*Server `json:"servers"`
	// }
	// We iterate over fields searching for the slice one to sort it
	resType := reflect.TypeOf(res).Elem()
	fields := reflect.VisibleFields(resType)
	for _, field := range fields {
		if field.Type.Kind() == reflect.Slice {
			sortSliceByZones(reflect.ValueOf(res).Elem().FieldByName(field.Name).Interface(), zones)
			return
		}
	}
}

// sortResponseByRegions find first field that is a slice in a struct and sort it by region
func sortResponseByRegions(res any, regions []Region) {
	// res may be ListServersResponse
	//
	// type ListServersResponse struct {
	//	TotalCount uint32 `json:"total_count"`
	//	Servers []*Server `json:"servers"`
	// }
	// We iterate over fields searching for the slice one to sort it
	resType := reflect.TypeOf(res).Elem()
	fields := reflect.VisibleFields(resType)
	for _, field := range fields {
		if field.Type.Kind() == reflect.Slice {
			sortSliceByRegions(reflect.ValueOf(res).Elem().FieldByName(field.Name).Interface(), regions)
			return
		}
	}
}

// newVariableFromType returns a variable set to the zero value of the given type
func newVariableFromType(t any) any {
	// reflect.New always create a pointer, that's why we use reflect.Indirect before
	return reflect.New(reflect.Indirect(reflect.ValueOf(t)).Type()).Interface()
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: http.DefaultTransport.(*http.Transport).Clone(),
	}
}

func setInsecureMode(c httpClient) {
	standardHTTPClient, ok := c.(*http.Client)
	if !ok {
		logger.Warningf("client: cannot use insecure mode with HTTP client of type %T", c)
		return
	}

	altTransport, ok := standardHTTPClient.Transport.(interface {
		SetInsecureTransport()
	})
	if ok {
		altTransport.SetInsecureTransport()
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

func setRequestLogging(c httpClient) {
	standardHTTPClient, ok := c.(*http.Client)
	if !ok {
		logger.Warningf("client: cannot use request logger with HTTP client of type %T", c)
		return
	}
	// Do not wrap transport if it is already a logger
	// As client is a pointer, changing transport will change given client
	// If the same httpClient is used in multiple scwClient, it would add multiple logger transports
	_, isLogger := standardHTTPClient.Transport.(*requestLoggerTransport)
	if !isLogger {
		standardHTTPClient.Transport = &requestLoggerTransport{rt: standardHTTPClient.Transport}
	}
}
