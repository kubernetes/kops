package hcloud

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
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
	ID   int64
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
	ID             int64
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
func (c *CertificateClient) GetByID(ctx context.Context, id int64) (*Certificate, *Response, error) {
	const opPath = "/certificates/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, id)

	respBody, resp, err := getRequest[schema.CertificateGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}
	return CertificateFromSchema(respBody.Certificate), resp, nil
}

// GetByName retrieves a Certificate by its name. If the Certificate does not exist, nil is returned.
func (c *CertificateClient) GetByName(ctx context.Context, name string) (*Certificate, *Response, error) {
	return firstByName(name, func() ([]*Certificate, *Response, error) {
		return c.List(ctx, CertificateListOpts{Name: name})
	})
}

// Get retrieves a Certificate by its ID if the input can be parsed as an integer, otherwise it
// retrieves a Certificate by its name. If the Certificate does not exist, nil is returned.
func (c *CertificateClient) Get(ctx context.Context, idOrName string) (*Certificate, *Response, error) {
	return getByIDOrName(ctx, c.GetByID, c.GetByName, idOrName)
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
	const opPath = "/certificates?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.values().Encode())

	respBody, resp, err := getRequest[schema.CertificateListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.Certificates, CertificateFromSchema), resp, nil
}

// All returns all Certificates.
func (c *CertificateClient) All(ctx context.Context) ([]*Certificate, error) {
	return c.AllWithOpts(ctx, CertificateListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all Certificates for the given options.
func (c *CertificateClient) AllWithOpts(ctx context.Context, opts CertificateListOpts) ([]*Certificate, error) {
	return iterPages(func(page int) ([]*Certificate, *Response, error) {
		opts.Page = page
		return c.List(ctx, opts)
	})
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
		return missingField(o, "Name")
	}
	switch o.Type {
	case "", CertificateTypeUploaded:
		return o.validateUploaded()
	case CertificateTypeManaged:
		return o.validateManaged()
	default:
		return invalidFieldValue(o, "Type", o.Type)
	}
}

func (o CertificateCreateOpts) validateManaged() error {
	if len(o.DomainNames) == 0 {
		return missingField(o, "DomainNames")
	}
	return nil
}

func (o CertificateCreateOpts) validateUploaded() error {
	if o.Certificate == "" {
		return missingField(o, "Certificate")
	}
	if o.PrivateKey == "" {
		return missingField(o, "PrivateKey")
	}
	return nil
}

// Create creates a new uploaded certificate.
//
// Create returns an error for certificates of any other type. Use
// CreateCertificate to create such certificates.
func (c *CertificateClient) Create(ctx context.Context, opts CertificateCreateOpts) (*Certificate, *Response, error) {
	if opts.Type != "" && opts.Type != CertificateTypeUploaded {
		return nil, nil, invalidFieldValue(opts, "Type", opts.Type)
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
	const opPath = "/certificates"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := opPath

	result := CertificateCreateResult{}

	if err := opts.Validate(); err != nil {
		return result, nil, err
	}

	reqBody := schema.CertificateCreateRequest{
		Name: opts.Name,
	}

	switch opts.Type {
	case "", CertificateTypeUploaded:
		reqBody.Type = string(CertificateTypeUploaded)
		reqBody.Certificate = opts.Certificate
		reqBody.PrivateKey = opts.PrivateKey
	case CertificateTypeManaged:
		reqBody.Type = string(CertificateTypeManaged)
		reqBody.DomainNames = opts.DomainNames
	default:
		return result, nil, invalidFieldValue(opts, "Type", opts.Type)
	}

	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}

	respBody, resp, err := postRequest[schema.CertificateCreateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return result, resp, err
	}

	result.Certificate = CertificateFromSchema(respBody.Certificate)
	if respBody.Action != nil {
		result.Action = ActionFromSchema(*respBody.Action)
	}

	return result, resp, nil
}

// CertificateUpdateOpts specifies options for updating a Certificate.
type CertificateUpdateOpts struct {
	Name   string
	Labels map[string]string
}

// Update updates a Certificate.
func (c *CertificateClient) Update(ctx context.Context, certificate *Certificate, opts CertificateUpdateOpts) (*Certificate, *Response, error) {
	const opPath = "/certificates/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, certificate.ID)

	reqBody := schema.CertificateUpdateRequest{}
	if opts.Name != "" {
		reqBody.Name = &opts.Name
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}

	respBody, resp, err := putRequest[schema.CertificateUpdateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return CertificateFromSchema(respBody.Certificate), resp, nil
}

// Delete deletes a certificate.
func (c *CertificateClient) Delete(ctx context.Context, certificate *Certificate) (*Response, error) {
	const opPath = "/certificates/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, certificate.ID)

	return deleteRequestNoResult(ctx, c.client, reqPath)
}

// RetryIssuance retries the issuance of a failed managed certificate.
func (c *CertificateClient) RetryIssuance(ctx context.Context, certificate *Certificate) (*Action, *Response, error) {
	const opPath = "/certificates/%d/actions/retry"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, certificate.ID)

	respBody, resp, err := postRequest[schema.CertificateIssuanceRetryResponse](ctx, c.client, reqPath, nil)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}
