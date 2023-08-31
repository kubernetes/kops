package hcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud/schema"
)

// CertificateType is the type of available certificate types.
type CertificateType string

// Available certificate types.
const (
	CertificateTypeUploaded CertificateType = "uploaded"
	CertificateTypeManaged  CertificateType = "managed"
)

// CertificateStatusType is defines the type for the various managed
// certificate status.
type CertificateStatusType string

// Possible certificate status.
const (
	CertificateStatusTypePending CertificateStatusType = "pending"
	CertificateStatusTypeFailed  CertificateStatusType = "failed"

	// only in issuance.
	CertificateStatusTypeCompleted CertificateStatusType = "completed"

	// only in renewal.
	CertificateStatusTypeScheduled   CertificateStatusType = "scheduled"
	CertificateStatusTypeUnavailable CertificateStatusType = "unavailable"
)

// CertificateUsedByRefType is the type of used by references for
// certificates.
type CertificateUsedByRefType string

// Possible users of certificates.
const (
	CertificateUsedByRefTypeLoadBalancer CertificateUsedByRefType = "load_balancer"
)

// CertificateUsedByRef points to a resource that uses this certificate.
type CertificateUsedByRef struct {
	ID   int
	Type CertificateUsedByRefType
}

// CertificateStatus indicates the status of a managed certificate.
type CertificateStatus struct {
	Issuance CertificateStatusType
	Renewal  CertificateStatusType
	Error    *Error
}

// IsFailed returns true if either the Issuance or the Renewal of a certificate
// failed. In this case the FailureReason field details the nature of the
// failure.
func (st *CertificateStatus) IsFailed() bool {
	return st.Issuance == CertificateStatusTypeFailed || st.Renewal == CertificateStatusTypeFailed
}

// Certificate represents a certificate in the Hetzner Cloud.
type Certificate struct {
	ID             int
	Name           string
	Labels         map[string]string
	Type           CertificateType
	Certificate    string
	Created        time.Time
	NotValidBefore time.Time
	NotValidAfter  time.Time
	DomainNames    []string
	Fingerprint    string
	Status         *CertificateStatus
	UsedBy         []CertificateUsedByRef
}

// CertificateCreateResult is the result of creating a certificate.
type CertificateCreateResult struct {
	Certificate *Certificate
	Action      *Action
}

// CertificateClient is a client for the Certificates API.
type CertificateClient struct {
	client *Client
	Action *ResourceActionClient
}

// GetByID retrieves a Certificate by its ID. If the Certificate does not exist, nil is returned.
func (c *CertificateClient) GetByID(ctx context.Context, id int) (*Certificate, *Response, error) {
	req, err := c.client.NewRequest(ctx, "GET", fmt.Sprintf("/certificates/%d", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.CertificateGetResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, nil, err
	}
	return CertificateFromSchema(body.Certificate), resp, nil
}

// GetByName retrieves a Certificate by its name. If the Certificate does not exist, nil is returned.
func (c *CertificateClient) GetByName(ctx context.Context, name string) (*Certificate, *Response, error) {
	if name == "" {
		return nil, nil, nil
	}
	Certificate, response, err := c.List(ctx, CertificateListOpts{Name: name})
	if len(Certificate) == 0 {
		return nil, response, err
	}
	return Certificate[0], response, err
}

// Get retrieves a Certificate by its ID if the input can be parsed as an integer, otherwise it
// retrieves a Certificate by its name. If the Certificate does not exist, nil is returned.
func (c *CertificateClient) Get(ctx context.Context, idOrName string) (*Certificate, *Response, error) {
	if id, err := strconv.Atoi(idOrName); err == nil {
		return c.GetByID(ctx, id)
	}
	return c.GetByName(ctx, idOrName)
}

// CertificateListOpts specifies options for listing Certificates.
type CertificateListOpts struct {
	ListOpts
	Name string
	Sort []string
}

func (l CertificateListOpts) values() url.Values {
	vals := l.ListOpts.Values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// List returns a list of Certificates for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *CertificateClient) List(ctx context.Context, opts CertificateListOpts) ([]*Certificate, *Response, error) {
	path := "/certificates?" + opts.values().Encode()
	req, err := c.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.CertificateListResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		return nil, nil, err
	}
	Certificates := make([]*Certificate, 0, len(body.Certificates))
	for _, s := range body.Certificates {
		Certificates = append(Certificates, CertificateFromSchema(s))
	}
	return Certificates, resp, nil
}

// All returns all Certificates.
func (c *CertificateClient) All(ctx context.Context) ([]*Certificate, error) {
	return c.AllWithOpts(ctx, CertificateListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all Certificates for the given options.
func (c *CertificateClient) AllWithOpts(ctx context.Context, opts CertificateListOpts) ([]*Certificate, error) {
	allCertificates := []*Certificate{}

	err := c.client.all(func(page int) (*Response, error) {
		opts.Page = page
		Certificates, resp, err := c.List(ctx, opts)
		if err != nil {
			return resp, err
		}
		allCertificates = append(allCertificates, Certificates...)
		return resp, nil
	})
	if err != nil {
		return nil, err
	}

	return allCertificates, nil
}

// CertificateCreateOpts specifies options for creating a new Certificate.
type CertificateCreateOpts struct {
	Name        string
	Type        CertificateType
	Certificate string
	PrivateKey  string
	Labels      map[string]string
	DomainNames []string
}

// Validate checks if options are valid.
func (o CertificateCreateOpts) Validate() error {
	if o.Name == "" {
		return errors.New("missing name")
	}
	switch o.Type {
	case "", CertificateTypeUploaded:
		return o.validateUploaded()
	case CertificateTypeManaged:
		return o.validateManaged()
	default:
		return fmt.Errorf("invalid type: %s", o.Type)
	}
}

func (o CertificateCreateOpts) validateManaged() error {
	if len(o.DomainNames) == 0 {
		return errors.New("no domain names")
	}
	return nil
}

func (o CertificateCreateOpts) validateUploaded() error {
	if o.Certificate == "" {
		return errors.New("missing certificate")
	}
	if o.PrivateKey == "" {
		return errors.New("missing private key")
	}
	return nil
}

// Create creates a new uploaded certificate.
//
// Create returns an error for certificates of any other type. Use
// CreateCertificate to create such certificates.
func (c *CertificateClient) Create(ctx context.Context, opts CertificateCreateOpts) (*Certificate, *Response, error) {
	if !(opts.Type == "" || opts.Type == CertificateTypeUploaded) {
		return nil, nil, fmt.Errorf("invalid certificate type: %s", opts.Type)
	}
	result, resp, err := c.CreateCertificate(ctx, opts)
	if err != nil {
		return nil, resp, err
	}
	return result.Certificate, resp, nil
}

// CreateCertificate creates a new certificate of any type.
func (c *CertificateClient) CreateCertificate(
	ctx context.Context, opts CertificateCreateOpts,
) (CertificateCreateResult, *Response, error) {
	var (
		action  *Action
		reqBody schema.CertificateCreateRequest
	)

	if err := opts.Validate(); err != nil {
		return CertificateCreateResult{}, nil, err
	}

	reqBody.Name = opts.Name

	switch opts.Type {
	case "", CertificateTypeUploaded:
		reqBody.Type = string(CertificateTypeUploaded)
		reqBody.Certificate = opts.Certificate
		reqBody.PrivateKey = opts.PrivateKey
	case CertificateTypeManaged:
		reqBody.Type = string(CertificateTypeManaged)
		reqBody.DomainNames = opts.DomainNames
	default:
		return CertificateCreateResult{}, nil, fmt.Errorf("invalid certificate type: %v", opts.Type)
	}

	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return CertificateCreateResult{}, nil, err
	}
	req, err := c.client.NewRequest(ctx, "POST", "/certificates", bytes.NewReader(reqBodyData))
	if err != nil {
		return CertificateCreateResult{}, nil, err
	}

	respBody := schema.CertificateCreateResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return CertificateCreateResult{}, resp, err
	}
	cert := CertificateFromSchema(respBody.Certificate)
	if respBody.Action != nil {
		action = ActionFromSchema(*respBody.Action)
	}

	return CertificateCreateResult{Certificate: cert, Action: action}, resp, nil
}

// CertificateUpdateOpts specifies options for updating a Certificate.
type CertificateUpdateOpts struct {
	Name   string
	Labels map[string]string
}

// Update updates a Certificate.
func (c *CertificateClient) Update(ctx context.Context, certificate *Certificate, opts CertificateUpdateOpts) (*Certificate, *Response, error) {
	reqBody := schema.CertificateUpdateRequest{}
	if opts.Name != "" {
		reqBody.Name = &opts.Name
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/certificates/%d", certificate.ID)
	req, err := c.client.NewRequest(ctx, "PUT", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	respBody := schema.CertificateUpdateResponse{}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return CertificateFromSchema(respBody.Certificate), resp, nil
}

// Delete deletes a certificate.
func (c *CertificateClient) Delete(ctx context.Context, certificate *Certificate) (*Response, error) {
	req, err := c.client.NewRequest(ctx, "DELETE", fmt.Sprintf("/certificates/%d", certificate.ID), nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req, nil)
}

// RetryIssuance retries the issuance of a failed managed certificate.
func (c *CertificateClient) RetryIssuance(ctx context.Context, certificate *Certificate) (*Action, *Response, error) {
	var respBody schema.CertificateIssuanceRetryResponse

	req, err := c.client.NewRequest(ctx, "POST", fmt.Sprintf("/certificates/%d/actions/retry", certificate.ID), nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, nil, err
	}
	action := ActionFromSchema(respBody.Action)
	return action, resp, nil
}
