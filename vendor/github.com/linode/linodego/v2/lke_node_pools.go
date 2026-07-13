package linodego

import (
	"context"
)

// LKELinodeStatus constants start with LKELinode and include
// Linode API LKENodePool Linode Status values
type LKELinodeStatus string

// LKENodePoolStatus constants reflect the current status of an LKENodePool
const (
	LKELinodeReady    LKELinodeStatus = "ready"
	LKELinodeNotReady LKELinodeStatus = "not_ready"
)

// LKENodePoolUpdateStrategy constants start with LKENodePool and include
// LKE Node Pool upgrade strategy values
type LKENodePoolUpdateStrategy string

// LKENodePoolUpdateStrategy constants describe the available upgrade strategies for LKE Enterprise only
const (
	LKENodePoolRollingUpdate LKENodePoolUpdateStrategy = "rolling_update"
	LKENodePoolOnRecycle     LKENodePoolUpdateStrategy = "on_recycle"
)

// LKENodePoolDisk represents a Node disk in an LKENodePool object
type LKENodePoolDisk struct {
	Size int    `json:"size"`
	Type string `json:"type"`
}

type LKENodePoolAutoscaler struct {
	Enabled bool `json:"enabled"`
	Min     int  `json:"min"`
	Max     int  `json:"max"`
}

// LKENodePoolLinode represents a LKENodePoolLinode object
type LKENodePoolLinode struct {
	ID         string          `json:"id"`
	InstanceID int             `json:"instance_id"`
	Status     LKELinodeStatus `json:"status"`
}

// LKENodePoolTaintEffect represents the effect value of a taint
type LKENodePoolTaintEffect string

const (
	LKENodePoolTaintEffectNoSchedule       LKENodePoolTaintEffect = "NoSchedule"
	LKENodePoolTaintEffectPreferNoSchedule LKENodePoolTaintEffect = "PreferNoSchedule"
	LKENodePoolTaintEffectNoExecute        LKENodePoolTaintEffect = "NoExecute"
)

// LKENodePoolTaint represents a corev1.Taint to add to an LKENodePool
type LKENodePoolTaint struct {
	Key    string                 `json:"key"`
	Value  string                 `json:"value,omitzero"`
	Effect LKENodePoolTaintEffect `json:"effect"`
}

// LKENodePoolIsolation controls network isolation for nodes in the pool.
type LKENodePoolIsolation struct {
	PublicIPv4 bool `json:"public_ipv4"`
	PublicIPv6 bool `json:"public_ipv6"`
}

// LKENodePoolIsolationCreateOptions controls network isolation for node pool create requests.
type LKENodePoolIsolationCreateOptions struct {
	PublicIPv4 *bool `json:"public_ipv4,omitzero"`
	PublicIPv6 *bool `json:"public_ipv6,omitzero"`
}

// LKENodePoolLabels represents Kubernetes labels to add to an LKENodePool
type LKENodePoolLabels map[string]string

// LKENodePool represents a LKENodePool object
type LKENodePool struct {
	ID      int                 `json:"id"`
	Count   int                 `json:"count"`
	Type    string              `json:"type"`
	Disks   []LKENodePoolDisk   `json:"disks"`
	Linodes []LKENodePoolLinode `json:"nodes"`
	Tags    []string            `json:"tags"`
	Labels  LKENodePoolLabels   `json:"labels"`
	Taints  []LKENodePoolTaint  `json:"taints"`
	Label   *string             `json:"label"`

	Autoscaler LKENodePoolAutoscaler `json:"autoscaler"`
	FirewallID *int                  `json:"firewall_id,omitzero"`

	DiskEncryption InstanceDiskEncryption `json:"disk_encryption,omitzero"`

	// Isolation may not currently be available to all users.
	Isolation *LKENodePoolIsolation `json:"isolation"`

	// K8sVersion and UpdateStrategy are only for LKE Enterprise to support node pool upgrades.
	// It may not currently be available to all users and is under v4beta.
	K8sVersion     *string                    `json:"k8s_version,omitzero"`
	UpdateStrategy *LKENodePoolUpdateStrategy `json:"update_strategy,omitzero"`

	// NOTE: Only cannot_delete applies to LKE node pools and can only be used with v4beta.
	Locks []LockType `json:"locks"`
}

// LKENodePoolCreateOptions fields are those accepted by CreateLKENodePool
type LKENodePoolCreateOptions struct {
	Count  int                `json:"count"`
	Type   string             `json:"type"`
	Disks  []LKENodePoolDisk  `json:"disks"`
	Tags   []string           `json:"tags"`
	Labels LKENodePoolLabels  `json:"labels"`
	Taints []LKENodePoolTaint `json:"taints"`
	Label  *string            `json:"label,omitzero"`

	Autoscaler *LKENodePoolAutoscaler `json:"autoscaler,omitzero"`
	FirewallID *int                   `json:"firewall_id,omitzero"`

	// Isolation may not currently be available to all users.
	Isolation *LKENodePoolIsolationCreateOptions `json:"isolation,omitzero"`

	// K8sVersion and UpdateStrategy only works for LKE Enterprise to support node pool upgrades.
	// It may not currently be available to all users and is under v4beta.
	K8sVersion     *string                    `json:"k8s_version,omitzero"`
	UpdateStrategy *LKENodePoolUpdateStrategy `json:"update_strategy,omitzero"`

	DiskEncryption *InstanceDiskEncryption `json:"disk_encryption,omitzero"`
}

// LKENodePoolUpdateOptions fields are those accepted by UpdateLKENodePoolUpdate
type LKENodePoolUpdateOptions struct {
	Count  int                `json:"count,omitzero"`
	Tags   []string           `json:"tags,omitzero"`
	Labels *LKENodePoolLabels `json:"labels,omitzero"`
	Taints []LKENodePoolTaint `json:"taints,omitzero"`
	Label  *string            `json:"label,omitzero"`

	Autoscaler *LKENodePoolAutoscaler `json:"autoscaler,omitzero"`
	FirewallID *int                   `json:"firewall_id,omitzero"`

	// K8sVersion and UpdateStrategy only works for LKE Enterprise to support node pool upgrades.
	// It may not currently be available to all users and is under v4beta.
	K8sVersion     *string                    `json:"k8s_version,omitzero"`
	UpdateStrategy *LKENodePoolUpdateStrategy `json:"update_strategy,omitzero"`
}

// GetCreateOptions converts a LKENodePool to LKENodePoolCreateOptions for
// use in CreateLKENodePool
func (l LKENodePool) GetCreateOptions() (o LKENodePoolCreateOptions) {
	o.Count = l.Count
	o.Disks = l.Disks
	o.Tags = l.Tags
	o.Labels = l.Labels
	o.Taints = l.Taints
	o.Autoscaler = &l.Autoscaler
	o.K8sVersion = l.K8sVersion
	o.UpdateStrategy = l.UpdateStrategy
	o.Label = l.Label
	o.FirewallID = l.FirewallID
	o.DiskEncryption = &l.DiskEncryption

	if l.Isolation != nil {
		o.Isolation = &LKENodePoolIsolationCreateOptions{
			PublicIPv4: Pointer(l.Isolation.PublicIPv4),
			PublicIPv6: Pointer(l.Isolation.PublicIPv6),
		}
	}

	return o
}

// GetUpdateOptions converts a LKENodePool to LKENodePoolUpdateOptions for use in UpdateLKENodePoolUpdate
func (l LKENodePool) GetUpdateOptions() (o LKENodePoolUpdateOptions) {
	o.Count = l.Count
	o.Tags = l.Tags
	o.Labels = &l.Labels
	o.Taints = l.Taints
	o.Autoscaler = &l.Autoscaler
	o.K8sVersion = l.K8sVersion
	o.UpdateStrategy = l.UpdateStrategy
	o.Label = l.Label
	o.FirewallID = l.FirewallID

	return o
}

// ListLKENodePools lists LKENodePools
func (c *Client) ListLKENodePools(ctx context.Context, clusterID int, opts *ListOptions) ([]LKENodePool, error) {
	return getPaginatedResults[LKENodePool](ctx, c, formatAPIPath("lke/clusters/%d/pools", clusterID), opts)
}

// GetLKENodePool gets the LKENodePool with the provided ID
func (c *Client) GetLKENodePool(ctx context.Context, clusterID, poolID int) (*LKENodePool, error) {
	e := formatAPIPath("lke/clusters/%d/pools/%d", clusterID, poolID)
	return doGETRequest[LKENodePool](ctx, c, e)
}

// CreateLKENodePool creates a LKENodePool
func (c *Client) CreateLKENodePool(ctx context.Context, clusterID int, opts LKENodePoolCreateOptions) (*LKENodePool, error) {
	e := formatAPIPath("lke/clusters/%d/pools", clusterID)
	return doPOSTRequest[LKENodePool](ctx, c, e, opts)
}

// RecycleLKENodePool recycles a LKENodePool
func (c *Client) RecycleLKENodePool(ctx context.Context, clusterID, poolID int) error {
	e := formatAPIPath("lke/clusters/%d/pools/%d/recycle", clusterID, poolID)
	return doPOSTRequestNoRequestResponseBody(ctx, c, e)
}

// UpdateLKENodePool updates the LKENodePool with the specified id
func (c *Client) UpdateLKENodePool(ctx context.Context, clusterID, poolID int, opts LKENodePoolUpdateOptions) (*LKENodePool, error) {
	e := formatAPIPath("lke/clusters/%d/pools/%d", clusterID, poolID)
	return doPUTRequest[LKENodePool](ctx, c, e, opts)
}

// DeleteLKENodePool deletes the LKENodePool with the specified id
func (c *Client) DeleteLKENodePool(ctx context.Context, clusterID, poolID int) error {
	e := formatAPIPath("lke/clusters/%d/pools/%d", clusterID, poolID)
	return doDELETERequest(ctx, c, e)
}

// GetLKENodePoolNode gets the LKENodePoolLinode with the provided ID
func (c *Client) GetLKENodePoolNode(ctx context.Context, clusterID int, nodeID string) (*LKENodePoolLinode, error) {
	e := formatAPIPath("lke/clusters/%d/nodes/%s", clusterID, nodeID)
	return doGETRequest[LKENodePoolLinode](ctx, c, e)
}

// RecycleLKENodePoolNode recycles a LKENodePoolLinode
func (c *Client) RecycleLKENodePoolNode(ctx context.Context, clusterID int, nodeID string) error {
	e := formatAPIPath("lke/clusters/%d/nodes/%s/recycle", clusterID, nodeID)
	return doPOSTRequestNoRequestResponseBody(ctx, c, e)
}

// DeleteLKENodePoolNode deletes a given node from a node pool
func (c *Client) DeleteLKENodePoolNode(ctx context.Context, clusterID int, nodeID string) error {
	e := formatAPIPath("lke/clusters/%d/nodes/%s", clusterID, nodeID)
	return doDELETERequest(ctx, c, e)
}
