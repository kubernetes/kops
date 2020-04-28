package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/jsonutil"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/uritemplates"
)

// Group defines a GCP Elastigroup.
type Group struct {
	ID          *string      `json:"id,omitempty"`
	Name        *string      `json:"name,omitempty"`
	Description *string      `json:"description,omitempty"`
	NodeImage   *string      `json:"nodeImage,omitempty"`
	Capacity    *Capacity    `json:"capacity,omitempty"`
	Compute     *Compute     `json:"compute,omitempty"`
	Scaling     *Scaling     `json:"scaling,omitempty"`
	Scheduling  *Scheduling  `json:"scheduling,omitempty"`
	Strategy    *Strategy    `json:"strategy,omitempty"`
	Integration *Integration `json:"thirdPartiesIntegration,omitempty"`

	// Read-only fields.
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`

	// forceSendFields is a list of field names (e.g. "Keys") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	forceSendFields []string

	// nullFields is a list of field names (e.g. "Keys") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	nullFields []string
}

// region AutoScale structs

type AutoScale struct {
	IsEnabled    *bool              `json:"isEnabled,omitempty"`
	IsAutoConfig *bool              `json:"isAutoConfig,omitempty"`
	Cooldown     *int               `json:"cooldown,omitempty"`
	Headroom     *AutoScaleHeadroom `json:"headroom,omitempty"`
	Down         *AutoScaleDown     `json:"down,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AutoScaleDown struct {
	EvaluationPeriods *int `json:"evaluationPeriods,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AutoScaleHeadroom struct {
	CPUPerUnit    *int `json:"cpuPerUnit,omitempty"`
	MemoryPerUnit *int `json:"memoryPerUnit,omitempty"`
	NumOfUnits    *int `json:"numOfUnits,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AutoScaleLabel struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// endregion

// region Capacity structs

// Capacity defines the capacity attributes of a Group instance
type Capacity struct {
	Maximum *int `json:"maximum,omitempty"`
	Minimum *int `json:"minimum,omitempty"`
	Target  *int `json:"target,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// endregion

// region Compute structs

// AccessConfig defines the access configuration for a network. AccessConfig is an element of NetworkInterface.
type AccessConfig struct {
	Name *string `json:"name,omitempty"`
	Type *string `json:"type,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// AliasIPRange defines the alias ip range for a network. AliasIPRange is an element of NetworkInterface.
type AliasIPRange struct {
	IPCIDRRange         *string `json:"ipCidrRange,omitempty"`
	SubnetworkRangeName *string `json:"subnetworkRangeName,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// BackendServiceConfig constains a list of backend service configurations.
type BackendServiceConfig struct {
	BackendServices []*BackendService `json:"backendServices,omitempty"`
	forceSendFields []string
	nullFields      []string
}

// BackendService defines the configuration for a single backend service.
type BackendService struct {
	BackendServiceName *string     `json:"backendServiceName,omitempty"`
	LocationType       *string     `json:"locationType,omitempty"`
	Scheme             *string     `json:"scheme,omitempty"`
	NamedPorts         *NamedPorts `json:"namedPorts,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// Compute defines the compute attributes of a Group.
type Compute struct {
	AvailabilityZones   []string             `json:"availabilityZones,omitempty"`
	GPU                 *GPU                 `json:"gpu,omitempty"`
	Health              *Health              `json:"health,omitempty"`
	InstanceTypes       *InstanceTypes       `json:"instanceTypes,omitempty"`
	LaunchSpecification *LaunchSpecification `json:"launchSpecification,omitempty"`
	Subnets             []*Subnet            `json:"subnets,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// CustomInstance defines the memory and vCPU constraints of an instance
type CustomInstance struct {
	VCPU      *int `json:"vCPU,omitempty"`
	MemoryGiB *int `json:"memoryGiB,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// Disk defines the a block of memory resources for the group. Stored in an array of Disks, as defined in LaunchSpecification.
type Disk struct {
	AutoDelete       *bool             `json:"autoDelete,omitempty"`
	Boot             *bool             `json:"boot,omitempty"`
	DeviceName       *string           `json:"deviceName,omitempty"`
	InitializeParams *InitializeParams `json:"initializeParams,omitempty"`
	Interface        *string           `json:"interface,omitempty"`
	Mode             *string           `json:"mode,omitempty"`
	Source           *string           `json:"source,omitempty"`
	Type             *string           `json:"type,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// GPU defines the kind and number of GPUs to use with the group. GPU is an element of Compute.
type GPU struct {
	Type  *string `json:"type,omitempty"`
	Count *int    `json:"count,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// Health defines the healthcheck attributes for the group. Health is an element of Compute.
type Health struct {
	AutoHealing       *bool   `json:"autoHealing,omitempty"`
	GracePeriod       *int    `json:"gracePeriod,omitempty"`
	HealthCheckType   *string `json:"healthCheckType,omitempty"`
	UnhealthyDuration *int    `json:"unhealthyDuration,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// InitializeParams defines the initialization parameters for a Disk object.
type InitializeParams struct {
	DiskSizeGB  *int    `json:"diskSizeGb,omitempty"`
	DiskType    *string `json:"diskType,omitempty"`
	SourceImage *string `json:"sourceImage,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// InstanceTypes defines the type of instances to use with the group. InstanceTypes is an element of Compute.
type InstanceTypes struct {
	OnDemand    *string           `json:"ondemand,omitempty"`
	Preemptible []string          `json:"preemptible,omitempty"`
	Custom      []*CustomInstance `json:"custom,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// Label defines an object holding a key:value pair. Label is an element of LaunchSpecification.
type Label struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// LaunchSpecification defines launch attributes for the Group. LaunchSpecification is an element of Compute.
type LaunchSpecification struct {
	BackendServiceConfig *BackendServiceConfig `json:"backendServiceConfig,omitempty"`
	Disks                []*Disk               `json:"disks,omitempty"`
	Labels               []*Label              `json:"labels,omitempty"`
	IPForwarding         *bool                 `json:"ipForwarding,omitempty"`
	NetworkInterfaces    []*NetworkInterface   `json:"networkInterfaces,omitempty"`
	Metadata             []*Metadata           `json:"metadata,omitempty"`
	ServiceAccount       *string               `json:"serviceAccount,omitempty"`
	StartupScript        *string               `json:"startupScript,omitempty"`
	ShutdownScript       *string               `json:"shutdownScript,omitempty"`
	Tags                 []string              `json:"tags,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// Metadata defines an object holding a key:value pair. Metadata is an element of LaunchSpecification.
type Metadata struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// NamedPorts describes the name and list of ports to use with the backend service
type NamedPorts struct {
	Name  *string `json:"name,omitempty"`
	Ports []int   `json:"ports,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// NetworkInterface defines the network configuration for a Group. NetworkInterface is an element of LaunchSpecification.
type NetworkInterface struct {
	AccessConfigs []*AccessConfig `json:"accessConfigs,omitempty"`
	AliasIPRanges []*AliasIPRange `json:"aliasIpRanges,omitempty"`
	Network       *string         `json:"network,omitempty"`
	ProjectID     *string         `json:"projectId,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// Subnet defines the attributes of a single Subnet. The Subnets list is an element of Compute.
type Subnet struct {
	Region      *string  `json:"region,omitempty"`
	SubnetNames []string `json:"subnetNames,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// endregion

// region GKE structs

// ImportGKEGroup contains a modified group struct used for overriding cluster parameters on import
type ImportGKEGroup struct {
	AvailabilityZones     []string          `json:"availabilityZones,omitempty"`
	Capacity              *CapacityGKE      `json:"capacity,omitempty"`
	Name                  *string           `json:"name,omitempty"`
	InstanceTypes         *InstanceTypesGKE `json:"instanceTypes,omitempty"`
	PreemptiblePercentage *int              `json:"preemptiblePercentage,omitempty"`
	NodeImage             *string           `json:"nodeImage,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type CapacityGKE struct {
	Capacity //embedding

	forceSendFields []string
	nullFields      []string
}

type InstanceTypesGKE struct {
	OnDemand    *string  `json:"ondemand,omitempty"`
	Preemptible []string `json:"preemptible,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// endregion

// region Scaling structs

// Action defines the action attributes of a ScalingPolicy.
type Action struct {
	Adjustment *int    `json:"adjustment,omitempty"`
	Type       *string `json:"type,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// Dimension defines the attributes for the dimensions of a ScalingPolicy.
type Dimension struct {
	Name  *string `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// Scaling defines the scaling attributes of a Group
type Scaling struct {
	Up   []*ScalingPolicy `json:"up,omitempty"`
	Down []*ScalingPolicy `json:"down,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// ScalingPolicy defines the scaling attributes for both up and down policies. ScalingPolicy is an element of Scaling.
type ScalingPolicy struct {
	Action            *Action      `json:"action,omitempty"`
	Cooldown          *int         `json:"cooldown,omitempty"`
	Dimensions        []*Dimension `json:"dimensions,omitempty"`
	EvaluationPeriods *int         `json:"evaluationPeriods,omitempty"`
	MetricName        *string      `json:"metricName,omitempty"`
	Namespace         *string      `json:"namespace,omitempty"`
	Operator          *string      `json:"operator,omitempty"`
	Period            *int         `json:"period,omitempty"`
	PolicyName        *string      `json:"policyName,omitempty"`
	Source            *string      `json:"source,omitempty"`
	Statistic         *string      `json:"statistic,omitempty"`
	Threshold         *float64     `json:"threshold,omitempty"`
	Unit              *string      `json:"unit,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// endregion

// region Strategy structs

// Strategy defines the strategy attributes of a Group.
type Strategy struct {
	DrainingTimeout       *int  `json:"drainingTimeout,omitempty"`
	FallbackToOnDemand    *bool `json:"fallbackToOd,omitempty"`
	PreemptiblePercentage *int  `json:"preemptiblePercentage,omitempty"`
	OnDemandCount         *int  `json:"onDemandCount,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// endregion

// region Scheduling

type Scheduling struct {
	Tasks []*Task `json:"tasks,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Task struct {
	IsEnabled      *bool   `json:"isEnabled,omitempty"`
	Type           *string `json:"taskType,omitempty"`
	CronExpression *string `json:"cronExpression,omitempty"`
	TargetCapacity *int    `json:"targetCapacity,omitempty"`
	MinCapacity    *int    `json:"minCapacity,omitempty"`
	MaxCapacity    *int    `json:"maxCapacity,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// endregion

// region Integration structs

type Integration struct {
	GKE         *GKEIntegration         `json:"gke,omitempty"`
	DockerSwarm *DockerSwarmIntegration `json:"dockerSwarm,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// region GKEIntegration structs

type GKEIntegration struct {
	ClusterID       *string       `json:"clusterIdentifier,omitempty"`
	ClusterZoneName *string       `json:"clusterZoneName,omitempty"`
	AutoUpdate      *bool         `json:"autoUpdate,omitempty"`
	AutoScale       *AutoScaleGKE `json:"autoScale,omitempty"`
	Location        *string       `json:"location,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AutoScaleGKE struct {
	AutoScale                   // embedding
	Labels    []*AutoScaleLabel `json:"labels,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// endregion

// region DockerSwarmIntegration structs

type DockerSwarmIntegration struct {
	MasterHost *string `json:"masterHost,omitempty"`
	MasterPort *int    `json:"masterPort,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// endregion

// endregion

// region API Operation structs

// CreateGroupInput contains the Elastigroup description required when making a request to create an Elastigroup.
type CreateGroupInput struct {
	Group *Group `json:"group,omitempty"`
}

// CreateGroupOutput contains a definition of the created Elastigroup, including the generated Group ID.
type CreateGroupOutput struct {
	Group *Group `json:"group,omitempty"`
}

// DeleteGroupInput contains the required input to delete an existing Elastigroup.
type DeleteGroupInput struct {
	GroupID *string `json:"groupId,omitempty"`
}

// DeleteGroupOutput describes the response a deleted group. Empty at this time.
type DeleteGroupOutput struct{}

// ImportGKEClusterInput describes the input required when importing an existing GKE cluster into Elastigroup, if it exists.
type ImportGKEClusterInput struct {
	ClusterID       *string         `json:"clusterID,omitempty"`
	ClusterZoneName *string         `json:"clusterZoneName,omitempty"`
	DryRun          *bool           `json:"dryRun,omitempty"`
	Group           *ImportGKEGroup `json:"group,omitempty"`
}

// ImportGKEClusterOutput contains a description of the Elastigroup and the imported GKE cluster.
type ImportGKEClusterOutput struct {
	Group *Group `json:"group,omitempty"`
}

// Instance describes an individual instance's status and is returned by a Status request
type Instance struct {
	CreatedAt    *time.Time `json:"createdAt,omitempty"`
	InstanceName *string    `json:"instanceName,omitempty"`
	LifeCycle    *string    `json:"lifeCycle,omitempty"`
	MachineType  *string    `json:"machineType,omitempty"`
	PrivateIP    *string    `json:"privateIpAddress,omitempty"`
	PublicIP     *string    `json:"publicIpAddress,omitempty"`
	StatusName   *string    `json:"statusName,omitempty"`
	UpdatedAt    *time.Time `json:"updatedAt,omitempty"`
	Zone         *string    `json:"zone,omitempty"`
}

// ListGroupsInput describes the input required when making a request to list all groups in an account.
type ListGroupsInput struct{}

// ListGroupsOutput contains an array of groups.
type ListGroupsOutput struct {
	Groups []*Group `json:"groups,omitempty"`
}

// ReadGroupInput describes the input required when making a request to list a single Elastigroup.
type ReadGroupInput struct {
	GroupID *string `json:"groupId,omitempty"`
}

// ReadGroupOutput contains a description of the requested Elastigroup, if it exists.
type ReadGroupOutput struct {
	Group *Group `json:"group,omitempty"`
}

// StatusGroupInput describes the required input when making a request to see an Elastigroup's status.
type StatusGroupInput struct {
	GroupID *string `json:"groupId,omitempty"`
}

// StatusGroupOutput describes the status of the instances in the Elastigroup.
type StatusGroupOutput struct {
	Instances []*Instance `json:"instances,omitempty"`
}

// UpdateGroupInput contains a description of one or more valid attributes that will be applied to an existing Elastigroup.
type UpdateGroupInput struct {
	Group *Group `json:"group,omitempty"`
}

// UpdateGroupOutPut contains a description of the updated Elastigroup, if successful.
type UpdateGroupOutput struct {
	Group *Group `json:"group,omitempty"`
}

// endregion

// region API Operations

// Create creates a new Elastigroup using GCE resources.
func (s *ServiceOp) Create(ctx context.Context, input *CreateGroupInput) (*CreateGroupOutput, error) {
	r := client.NewRequest(http.MethodPost, "/gcp/gce/group")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := groupsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateGroupOutput)
	if len(gs) > 0 {
		output.Group = gs[0]
	}

	return output, nil
}

// Read returns the configuration of a single existing Elastigroup.
func (s *ServiceOp) Read(ctx context.Context, input *ReadGroupInput) (*ReadGroupOutput, error) {
	path, err := uritemplates.Expand("/gcp/gce/group/{groupId}", uritemplates.Values{
		"groupId": spotinst.StringValue(input.GroupID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodGet, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := groupsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadGroupOutput)
	if len(gs) > 0 {
		output.Group = gs[0]
	}

	return output, nil
}

// Update modifies the configuration of a single existing Elastigroup.
func (s *ServiceOp) Update(ctx context.Context, input *UpdateGroupInput) (*UpdateGroupOutput, error) {
	path, err := uritemplates.Expand("/gcp/gce/group/{groupId}", uritemplates.Values{
		"groupId": spotinst.StringValue(input.Group.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do NOT need the ID anymore, so let's drop it.
	input.Group.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := groupsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(UpdateGroupOutput)
	if len(gs) > 0 {
		output.Group = gs[0]
	}

	return output, nil
}

// Delete removes a single existing Elastigroup and destroys all associated GCE resources.
func (s *ServiceOp) Delete(ctx context.Context, input *DeleteGroupInput) (*DeleteGroupOutput, error) {
	path, err := uritemplates.Expand("/gcp/gce/group/{groupId}", uritemplates.Values{
		"groupId": spotinst.StringValue(input.GroupID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodDelete, path)

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DeleteGroupOutput{}, nil
}

// List returns the configuration of all existing Elastigroups in a given Spotinst GCE account.
func (s *ServiceOp) List(ctx context.Context, input *ListGroupsInput) (*ListGroupsOutput, error) {
	r := client.NewRequest(http.MethodGet, "/gcp/gce/group")
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := groupsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListGroupsOutput{Groups: gs}, nil
}

// ImportGKECluster imports an existing GKE cluster into Elastigroup.
func (s *ServiceOp) ImportGKECluster(ctx context.Context, input *ImportGKEClusterInput) (*ImportGKEClusterOutput, error) {
	r := client.NewRequest(http.MethodPost, "/gcp/gce/group/gke/import")

	r.Params["clusterId"] = []string{spotinst.StringValue(input.ClusterID)}
	r.Params["zone"] = []string{spotinst.StringValue(input.ClusterZoneName)}
	r.Params["dryRun"] = []string{strconv.FormatBool(spotinst.BoolValue(input.DryRun))}

	body := &ImportGKEClusterInput{Group: input.Group}
	r.Obj = body

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := groupsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ImportGKEClusterOutput)
	if len(gs) > 0 {
		output.Group = gs[0]
	}

	return output, nil
}

// Status describes the current status of the instances in a specific Elastigroup
func (s *ServiceOp) Status(ctx context.Context, input *StatusGroupInput) (*StatusGroupOutput, error) {
	path, err := uritemplates.Expand("/gcp/gce/group/{groupId}/status", uritemplates.Values{
		"groupId": spotinst.StringValue(input.GroupID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodGet, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	is, err := instancesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &StatusGroupOutput{Instances: is}, nil
}

// endregion

// region Unmarshallers

// groupFromJSON unmarshalls a single group
func groupFromJSON(in []byte) (*Group, error) {
	b := new(Group)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

// groupsFromJSON unmarshalls an array of groups
func groupsFromJSON(in []byte) ([]*Group, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Group, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := groupFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

// groupFromJSON reads a list of one or more groups from an http response
func groupsFromHttpResponse(resp *http.Response) ([]*Group, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return groupsFromJSON(body)
}

// instanceFromJSON unmarshalls a single group
func instanceFromJSON(in []byte) (*Instance, error) {
	b := new(Instance)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

// instancesFromJSON unmarshalls an array of instances
func instancesFromJSON(in []byte) ([]*Instance, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Instance, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := instanceFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

// instancesFromHttpResponse reads a list of one or more instances from an http response
func instancesFromHttpResponse(resp *http.Response) ([]*Instance, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return instancesFromJSON(body)
}

// endregion

// region Group setters

func (o Group) MarshalJSON() ([]byte, error) {
	type noMethod Group
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetID sets the group ID attribute
func (o *Group) SetID(v *string) *Group {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

// SetName sets the group name
func (o *Group) SetName(v *string) *Group {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

// SetDescription sets the description for the group
func (o *Group) SetDescription(v *string) *Group {
	if o.Description = v; o.Description == nil {
		o.nullFields = append(o.nullFields, "Description")
	}
	return o
}

// SetNodeImage sets image that will be used for the node VMs
func (o *Group) SetNodeImage(v *string) *Group {
	if o.NodeImage = v; o.NodeImage == nil {
		o.nullFields = append(o.nullFields, "NodeImage")
	}
	return o
}

// SetCapacity sets the Capacity object
func (o *Group) SetCapacity(v *Capacity) *Group {
	if o.Capacity = v; o.Capacity == nil {
		o.nullFields = append(o.nullFields, "Capacity")
	}
	return o
}

// SetCompute sets the Compute object
func (o *Group) SetCompute(v *Compute) *Group {
	if o.Compute = v; o.Compute == nil {
		o.nullFields = append(o.nullFields, "Compute")
	}
	return o
}

// SetScaling sets the Scaling object
func (o *Group) SetScaling(v *Scaling) *Group {
	if o.Scaling = v; o.Scaling == nil {
		o.nullFields = append(o.nullFields, "Scaling")
	}
	return o
}

func (o *Group) SetScheduling(v *Scheduling) *Group {
	if o.Scheduling = v; o.Scheduling == nil {
		o.nullFields = append(o.nullFields, "Scheduling")
	}
	return o
}

// SetStrategy sets the Strategy object
func (o *Group) SetStrategy(v *Strategy) *Group {
	if o.Strategy = v; o.Strategy == nil {
		o.nullFields = append(o.nullFields, "Strategy")
	}
	return o
}

// SetIntegration sets the integrations for the group
func (o *Group) SetIntegration(v *Integration) *Group {
	if o.Integration = v; o.Integration == nil {
		o.nullFields = append(o.nullFields, "Integration")
	}
	return o
}

// endregion

// region AutoScale setters

func (o AutoScale) MarshalJSON() ([]byte, error) {
	type noMethod AutoScale
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *AutoScale) SetIsEnabled(v *bool) *AutoScale {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}

func (o *AutoScale) SetIsAutoConfig(v *bool) *AutoScale {
	if o.IsAutoConfig = v; o.IsAutoConfig == nil {
		o.nullFields = append(o.nullFields, "IsAutoConfig")
	}
	return o
}

func (o *AutoScale) SetCooldown(v *int) *AutoScale {
	if o.Cooldown = v; o.Cooldown == nil {
		o.nullFields = append(o.nullFields, "Cooldown")
	}
	return o
}

func (o *AutoScale) SetHeadroom(v *AutoScaleHeadroom) *AutoScale {
	if o.Headroom = v; o.Headroom == nil {
		o.nullFields = append(o.nullFields, "Headroom")
	}
	return o
}

func (o *AutoScale) SetDown(v *AutoScaleDown) *AutoScale {
	if o.Down = v; o.Down == nil {
		o.nullFields = append(o.nullFields, "Down")
	}
	return o
}

// region AutoScaleDown

func (o AutoScaleDown) MarshalJSON() ([]byte, error) {
	type noMethod AutoScaleDown
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *AutoScaleDown) SetEvaluationPeriods(v *int) *AutoScaleDown {
	if o.EvaluationPeriods = v; o.EvaluationPeriods == nil {
		o.nullFields = append(o.nullFields, "EvaluationPeriods")
	}
	return o
}

// endregion

// region AutoScaleHeadroom

func (o AutoScaleHeadroom) MarshalJSON() ([]byte, error) {
	type noMethod AutoScaleHeadroom
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *AutoScaleHeadroom) SetCPUPerUnit(v *int) *AutoScaleHeadroom {
	if o.CPUPerUnit = v; o.CPUPerUnit == nil {
		o.nullFields = append(o.nullFields, "CPUPerUnit")
	}
	return o
}

func (o *AutoScaleHeadroom) SetMemoryPerUnit(v *int) *AutoScaleHeadroom {
	if o.MemoryPerUnit = v; o.MemoryPerUnit == nil {
		o.nullFields = append(o.nullFields, "MemoryPerUnit")
	}
	return o
}

func (o *AutoScaleHeadroom) SetNumOfUnits(v *int) *AutoScaleHeadroom {
	if o.NumOfUnits = v; o.NumOfUnits == nil {
		o.nullFields = append(o.nullFields, "NumOfUnits")
	}
	return o
}

// endregion

// region AutoScaleLabel

func (o AutoScaleLabel) MarshalJSON() ([]byte, error) {
	type noMethod AutoScaleLabel
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *AutoScaleLabel) SetKey(v *string) *AutoScaleLabel {
	if o.Key = v; o.Key == nil {
		o.nullFields = append(o.nullFields, "Key")
	}
	return o
}

func (o *AutoScaleLabel) SetValue(v *string) *AutoScaleLabel {
	if o.Value = v; o.Value == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}

// endregion

// endregion

// region Capacity setters

func (o Capacity) MarshalJSON() ([]byte, error) {
	type noMethod Capacity
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetMaximum sets the Maximum number of VMs in the group.
func (o *Capacity) SetMaximum(v *int) *Capacity {
	if o.Maximum = v; o.Maximum == nil {
		o.nullFields = append(o.nullFields, "Maximum")
	}
	return o
}

// SetMinimum sets the minimum number of VMs in the group
func (o *Capacity) SetMinimum(v *int) *Capacity {
	if o.Minimum = v; o.Minimum == nil {
		o.nullFields = append(o.nullFields, "Minimum")
	}
	return o
}

// SetTarget sets the desired number of running VMs in the group.
func (o *Capacity) SetTarget(v *int) *Capacity {
	if o.Target = v; o.Target == nil {
		o.nullFields = append(o.nullFields, "Target")
	}
	return o
}

// endregion

// region Compute setters

func (o Compute) MarshalJSON() ([]byte, error) {
	type noMethod Compute
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetAvailabilityZones sets the list of availability zones for group resources.
func (o *Compute) SetAvailabilityZones(v []string) *Compute {
	if o.AvailabilityZones = v; o.AvailabilityZones == nil {
		o.nullFields = append(o.nullFields, "AvailabilityZones")
	}
	return o
}

// SetGPU sets the GPU object
func (o *Compute) SetGPU(v *GPU) *Compute {
	if o.GPU = v; o.GPU == nil {
		o.nullFields = append(o.nullFields, "GPU")
	}
	return o
}

// SetHealth sets the health check attributes for the group
func (o *Compute) SetHealth(v *Health) *Compute {
	if o.Health = v; o.Health == nil {
		o.nullFields = append(o.nullFields, "Health")
	}
	return o
}

// SetInstanceTypes sets the instance types for the group.
func (o *Compute) SetInstanceTypes(v *InstanceTypes) *Compute {
	if o.InstanceTypes = v; o.InstanceTypes == nil {
		o.nullFields = append(o.nullFields, "InstanceTypes")
	}
	return o
}

// SetLaunchSpecification sets the launch configuration of the group.
func (o *Compute) SetLaunchConfiguration(v *LaunchSpecification) *Compute {
	if o.LaunchSpecification = v; o.LaunchSpecification == nil {
		o.nullFields = append(o.nullFields, "LaunchSpecification")
	}
	return o
}

// SetSubnets sets the subnets used by the group.
func (o *Compute) SetSubnets(v []*Subnet) *Compute {
	if o.Subnets = v; o.Subnets == nil {
		o.nullFields = append(o.nullFields, "Subnets")
	}
	return o
}

// region GPU Setters

func (o GPU) MarshalJSON() ([]byte, error) {
	type noMethod GPU
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetType sets the type of gpu
func (o *GPU) SetType(v *string) *GPU {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

// SetCount sets the number of this type of gpu
func (o *GPU) SetCount(v *int) *GPU {
	if o.Count = v; o.Count == nil {
		o.nullFields = append(o.nullFields, "Count")
	}
	return o
}

// endregion

// region Health setters

func (o Health) MarshalJSON() ([]byte, error) {
	type noMethod Health
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetGracePeriod sets the grace period time for the groups health check
func (o *Health) SetGracePeriod(v *int) *Health {
	fmt.Printf("o: %v\n", o)
	if o.GracePeriod = v; o.GracePeriod == nil {
		o.nullFields = append(o.nullFields, "GracePeriod")
	}
	return o
}

// SetHealthCheckType sets the type of helath check to perform
func (o *Health) SetHealthCheckType(v *string) *Health {
	if o.HealthCheckType = v; o.HealthCheckType == nil {
		o.nullFields = append(o.nullFields, "HealthCheckType")
	}
	return o
}

// SetAutoHealing sets autohealing to true or false
func (o *Health) SetAutoHealing(v *bool) *Health {
	if o.AutoHealing = v; o.AutoHealing == nil {
		o.nullFields = append(o.nullFields, "AutoHealing")
	}
	return o
}

func (o *Health) SetUnhealthyDuration(v *int) *Health {
	if o.UnhealthyDuration = v; o.UnhealthyDuration == nil {
		o.nullFields = append(o.nullFields, "UnhealthyDuration")
	}
	return o
}

// endregion

// region InstanceTypes setters

func (o InstanceTypes) MarshalJSON() ([]byte, error) {
	type noMethod InstanceTypes
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetCustom sets the memory and vCPU attributes for Custom Instance types
func (o *InstanceTypes) SetCustom(v []*CustomInstance) *InstanceTypes {
	if o.Custom = v; o.Custom == nil {
		o.nullFields = append(o.nullFields, "Custom")
	}
	return o
}

// SetMemoryGiB sets the memory amount for a Custom Instance in intervals of 2, min 10
func (o *CustomInstance) SetMemoryGiB(v *int) *CustomInstance {
	if o.MemoryGiB = v; o.MemoryGiB == nil {
		o.nullFields = append(o.nullFields, "MemoryGiB")
	}
	return o
}

// SetVCPU sets sets the number of vCPUs to use in a Custom instance type
func (o *CustomInstance) SetVCPU(v *int) *CustomInstance {
	if o.VCPU = v; o.VCPU == nil {
		o.nullFields = append(o.nullFields, "VCPU")
	}
	return o
}

// SetOnDemand sets the kind of on demand instances to use for the group.
func (o *InstanceTypes) SetOnDemand(v *string) *InstanceTypes {
	if o.OnDemand = v; o.OnDemand == nil {
		o.nullFields = append(o.nullFields, "OnDemand")
	}
	return o
}

// SetPreemptible sets the kind of premeptible instances to use with the group.
func (o *InstanceTypes) SetPreemptible(v []string) *InstanceTypes {
	if o.Preemptible = v; o.Preemptible == nil {
		o.nullFields = append(o.nullFields, "Preemptible")
	}
	return o
}

// endregion

// region LaunchSpecification setters

func (o LaunchSpecification) MarshalJSON() ([]byte, error) {
	type noMethod LaunchSpecification
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetBackendServices sets the backend services to use with the group.
func (o *LaunchSpecification) SetBackendServiceConfig(v *BackendServiceConfig) *LaunchSpecification {
	if o.BackendServiceConfig = v; o.BackendServiceConfig == nil {
		o.nullFields = append(o.nullFields, "BackendServiceConfig")
	}
	return o
}

// SetDisks sets the list of disks used by the group
func (o *LaunchSpecification) SetDisks(v []*Disk) *LaunchSpecification {
	if o.Disks = v; o.Disks == nil {
		o.nullFields = append(o.nullFields, "Disks")
	}
	return o
}

// SetLabels sets the labels to be used with the group
func (o *LaunchSpecification) SetLabels(v []*Label) *LaunchSpecification {
	if o.Labels = v; o.Labels == nil {
		o.nullFields = append(o.nullFields, "Labels")
	}
	return o
}

// SetIPForwarding sets whether to use IP forwarding for this group.
func (o *LaunchSpecification) SetIPForwarding(v *bool) *LaunchSpecification {
	if o.IPForwarding = v; o.IPForwarding == nil {
		o.nullFields = append(o.nullFields, "IPForwarding")
	}
	return o
}

// SetNetworkInterfaces sets number and kinds of network interfaces used by the group.
func (o *LaunchSpecification) SetNetworkInterfaces(v []*NetworkInterface) *LaunchSpecification {
	if o.NetworkInterfaces = v; o.NetworkInterfaces == nil {
		o.nullFields = append(o.nullFields, "NetworkInterfaces")
	}
	return o
}

// SetMetadata sets metadata for the group.
func (o *LaunchSpecification) SetMetadata(v []*Metadata) *LaunchSpecification {
	if o.Metadata = v; o.Metadata == nil {
		o.nullFields = append(o.nullFields, "Metadata")
	}
	return o
}

// SetServiceAccount sets the service account used by the instances in the group
func (o *LaunchSpecification) SetServiceAccount(v *string) *LaunchSpecification {
	if o.ServiceAccount = v; o.ServiceAccount == nil {
		o.nullFields = append(o.nullFields, "ServiceAccount")
	}
	return o
}

// SetStartupScript sets the startup script to be executed when the instance launches.
func (o *LaunchSpecification) SetStartupScript(v *string) *LaunchSpecification {
	if o.StartupScript = v; o.StartupScript == nil {
		o.nullFields = append(o.nullFields, "StartupScript")
	}
	return o
}

// SetShutdownScript sets the script that will run when draining instances before termination
func (o *LaunchSpecification) SetShutdownScript(v *string) *LaunchSpecification {
	if o.ShutdownScript = v; o.ShutdownScript == nil {
		o.nullFields = append(o.nullFields, "ShutdownScript")
	}
	return o
}

// SetTags sets the list of tags
func (o *LaunchSpecification) SetTags(v []string) *LaunchSpecification {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

// region BackendServiceConfig setters

func (o BackendServiceConfig) MarshalJSON() ([]byte, error) {
	type noMethod BackendServiceConfig
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetBackendServices sets the backend service list
func (o *BackendServiceConfig) SetBackendServices(v []*BackendService) *BackendServiceConfig {
	if o.BackendServices = v; o.BackendServices == nil {
		o.nullFields = append(o.nullFields, "BackendServices")
	}
	return o
}

// region Backend Service setters

func (o BackendService) MarshalJSON() ([]byte, error) {
	type noMethod BackendService
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetBackendServiceName sets the name of the backend service.
func (o *BackendService) SetBackendServiceName(v *string) *BackendService {
	if o.BackendServiceName = v; o.BackendServiceName == nil {
		o.nullFields = append(o.nullFields, "BackendServiceName")
	}
	return o
}

// SetLocationType sets the location type
func (o *BackendService) SetLocationType(v *string) *BackendService {
	if o.LocationType = v; o.LocationType == nil {
		o.nullFields = append(o.nullFields, "LocationType")
	}
	return o
}

// SetScheme sets the scheme
func (o *BackendService) SetScheme(v *string) *BackendService {
	if o.Scheme = v; o.Scheme == nil {
		o.nullFields = append(o.nullFields, "Scheme")
	}
	return o
}

// SetNamedPorts sets the named port object
func (o *BackendService) SetNamedPorts(v *NamedPorts) *BackendService {
	if o.NamedPorts = v; o.NamedPorts == nil {
		o.nullFields = append(o.nullFields, "NamedPort")
	}
	return o
}

// region NamedPort setters

func (o NamedPorts) MarshalJSON() ([]byte, error) {
	type noMethod NamedPorts
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetNamedPorts sets the name of the NamedPorts
func (o *NamedPorts) SetName(v *string) *NamedPorts {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

// SetPorts sets the list of ports in the NamedPorts
func (o *NamedPorts) SetPorts(v []int) *NamedPorts {
	if o.Ports = v; o.Ports == nil {
		o.nullFields = append(o.nullFields, "Ports")
	}
	return o
}

// endregion

// endregion

// endregion

// region Disk setters

func (o Disk) MarshalJSON() ([]byte, error) {
	type noMethod Disk
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetAutoDelete sets option to have disks autodelete
func (o *Disk) SetAutoDelete(v *bool) *Disk {
	if o.AutoDelete = v; o.AutoDelete == nil {
		o.nullFields = append(o.nullFields, "AutoDelete")
	}
	return o
}

// SetBoot sets the boot option
func (o *Disk) SetBoot(v *bool) *Disk {
	if o.Boot = v; o.Boot == nil {
		o.nullFields = append(o.nullFields, "Boot")
	}
	return o
}

// SetDeviceName sets the device name
func (o *Disk) SetDeviceName(v *string) *Disk {
	if o.DeviceName = v; o.DeviceName == nil {
		o.nullFields = append(o.nullFields, "DeviceName")
	}
	return o
}

// SetInitializeParams sets the initialization paramters object
func (o *Disk) SetInitializeParams(v *InitializeParams) *Disk {
	if o.InitializeParams = v; o.InitializeParams == nil {
		o.nullFields = append(o.nullFields, "InitializeParams")
	}
	return o
}

// SetInterface sets the interface
func (o *Disk) SetInterface(v *string) *Disk {
	if o.Interface = v; o.Interface == nil {
		o.nullFields = append(o.nullFields, "Interface")
	}
	return o
}

// SetMode sets the mode
func (o *Disk) SetMode(v *string) *Disk {
	if o.Mode = v; o.Mode == nil {
		o.nullFields = append(o.nullFields, "Mode")
	}
	return o
}

// SetSource sets the source
func (o *Disk) SetSource(v *string) *Disk {
	if o.Source = v; o.Source == nil {
		o.nullFields = append(o.nullFields, "Source")
	}
	return o
}

// SetType sets the type of disk
func (o *Disk) SetType(v *string) *Disk {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

// region InitializeParams setters

func (o InitializeParams) MarshalJSON() ([]byte, error) {
	type noMethod InitializeParams
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetDiskSizeGB sets the disk size in gigabytes, in multiples of 2
func (o *InitializeParams) SetDiskSizeGB(v *int) *InitializeParams {
	if o.DiskSizeGB = v; o.DiskSizeGB == nil {
		o.nullFields = append(o.nullFields, "DiskSizeGB")
	}
	return o
}

// SetDiskType sets the type of disk
func (o *InitializeParams) SetDiskType(v *string) *InitializeParams {
	if o.DiskType = v; o.DiskType == nil {
		o.nullFields = append(o.nullFields, "DiskType")
	}
	return o
}

// SetSourceImage sets the source image to use
func (o *InitializeParams) SetSourceImage(v *string) *InitializeParams {
	if o.SourceImage = v; o.SourceImage == nil {
		o.nullFields = append(o.nullFields, "SourceImage")
	}
	return o
}

// endregion

// endregion

// region Label setters

func (o Label) MarshalJSON() ([]byte, error) {
	type noMethod Label
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetKey sets the key for the label
func (o *Label) SetKey(v *string) *Label {
	if o.Key = v; o.Key == nil {
		o.nullFields = append(o.nullFields, "Key")
	}
	return o
}

// SetValue sets the value for the label
func (o *Label) SetValue(v *string) *Label {
	if o.Value = v; o.Value == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}

// endregion

// region NetworkInterface setters

func (o NetworkInterface) MarshalJSON() ([]byte, error) {
	type noMethod NetworkInterface
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetAccessConfigs creates a list of one or more access configuration objects
func (o *NetworkInterface) SetAccessConfigs(v []*AccessConfig) *NetworkInterface {
	if o.AccessConfigs = v; o.AccessConfigs == nil {
		o.nullFields = append(o.nullFields, "AccessConfigs")
	}
	return o
}

// SetAliasIPRanges sets a list of alias IP range objects
func (o *NetworkInterface) SetAliasIPRanges(v []*AliasIPRange) *NetworkInterface {
	if o.AliasIPRanges = v; o.AliasIPRanges == nil {
		o.nullFields = append(o.nullFields, "AliasIPRanges")
	}
	return o
}

// SetNetwork sets the name of the network
func (o *NetworkInterface) SetNetwork(v *string) *NetworkInterface {
	if o.Network = v; o.Network == nil {
		o.nullFields = append(o.nullFields, "Network")
	}
	return o
}

// SetProjectId sets the project identifier of the network.
func (o *NetworkInterface) SetProjectId(v *string) *NetworkInterface {
	if o.ProjectID = v; o.ProjectID == nil {
		o.nullFields = append(o.nullFields, "ProjectID")
	}
	return o
}

// region AccessConfig setters

func (o AccessConfig) MarshalJSON() ([]byte, error) {
	type noMethod AccessConfig
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetName sets the name of the access configuration
func (o *AccessConfig) SetName(v *string) *AccessConfig {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

// SetType sets the type of access configuration
func (o *AccessConfig) SetType(v *string) *AccessConfig {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

// endregion

// region AliasIPRange setters

func (o AliasIPRange) MarshalJSON() ([]byte, error) {
	type noMethod AliasIPRange
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetIPCIDRRange sets the ip/cidr range
func (o *AliasIPRange) SetIPCIDRRange(v *string) *AliasIPRange {
	if o.IPCIDRRange = v; o.IPCIDRRange == nil {
		o.nullFields = append(o.nullFields, "IPCIDRRange")
	}
	return o
}

// SetSubnetworkRangeName sets the name of the subnetwork range
func (o *AliasIPRange) SetSubnetworkRangeName(v *string) *AliasIPRange {
	if o.SubnetworkRangeName = v; o.SubnetworkRangeName == nil {
		o.nullFields = append(o.nullFields, "SubnetworkRangeName")
	}
	return o
}

// endregion

// endregion

// region Metadata setters

func (o Metadata) MarshalJSON() ([]byte, error) {
	type noMethod Metadata
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetKey sets the metadata key
func (o *Metadata) SetKey(v *string) *Metadata {
	if o.Key = v; o.Key == nil {
		o.nullFields = append(o.nullFields, "Key")
	}
	return o
}

// SetValue sets the metadata value
func (o *Metadata) SetValue(v *string) *Metadata {
	if o.Value = v; o.Value == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}

// endregion

// endregion

// region Subnet setters

func (o Subnet) MarshalJSON() ([]byte, error) {
	type noMethod Subnet
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetRegion sets the region the subnet is in.
func (o *Subnet) SetRegion(v *string) *Subnet {
	if o.Region = v; o.Region == nil {
		o.nullFields = append(o.nullFields, "Region")
	}
	return o
}

// SetSubnetNames sets the list of subnets names to use
func (o *Subnet) SetSubnetNames(v []string) *Subnet {
	if o.SubnetNames = v; o.SubnetNames == nil {
		o.nullFields = append(o.nullFields, "SubnetNames")
	}
	return o
}

// endregion

// endregion

// region ImportGKE setters

func (o ImportGKEGroup) MarshalJSON() ([]byte, error) {
	type noMethod ImportGKEGroup
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetAvailabilityZones sets the availability zones for the gke group
func (o *ImportGKEGroup) SetAvailabilityZones(v []string) *ImportGKEGroup {
	if o.AvailabilityZones = v; o.AvailabilityZones == nil {
		o.nullFields = append(o.nullFields, "AvailabilityZones")
	}
	return o
}

// SetCapacity sets the capacity for a gke group
func (o *ImportGKEGroup) SetCapacity(v *CapacityGKE) *ImportGKEGroup {
	if o.Capacity = v; o.Capacity == nil {
		o.nullFields = append(o.nullFields, "Capacity")
	}
	return o
}

// SetInstanceTypes sets the instance types for the group.
func (o *ImportGKEGroup) SetInstanceTypes(v *InstanceTypesGKE) *ImportGKEGroup {
	if o.InstanceTypes = v; o.InstanceTypes == nil {
		o.nullFields = append(o.nullFields, "InstanceTypes")
	}
	return o
}

// SetName sets the group name
func (o *ImportGKEGroup) SetName(v *string) *ImportGKEGroup {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

// SetPreemptiblePercentage sets the preemptible percentage when importing a gke cluster into Elastigroup.
func (o *ImportGKEGroup) SetPreemptiblePercentage(v *int) *ImportGKEGroup {
	if o.PreemptiblePercentage = v; o.PreemptiblePercentage == nil {
		o.nullFields = append(o.nullFields, "PreemptiblePercentage")
	}
	return o
}

// SetNodeImage sets the node image for the imported gke group.
func (o *ImportGKEGroup) SetNodeImage(v *string) *ImportGKEGroup {
	if o.NodeImage = v; o.NodeImage == nil {
		o.nullFields = append(o.nullFields, "NodeImage")
	}
	return o
}

func (o InstanceTypesGKE) MarshalJSON() ([]byte, error) {
	type noMethod InstanceTypesGKE
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetOnDemand sets the instance types when importing a gke group
func (o *InstanceTypesGKE) SetOnDemand(v *string) *InstanceTypesGKE {
	if o.OnDemand = v; o.OnDemand == nil {
		o.nullFields = append(o.nullFields, "OnDemand")
	}
	return o
}

// SetPreemptible sets the list of preemptible instance types
func (o *InstanceTypesGKE) SetPreemptible(v []string) *InstanceTypesGKE {
	if o.Preemptible = v; o.Preemptible == nil {
		o.nullFields = append(o.nullFields, "Preemptible")
	}
	return o
}

// endregion

// region Integration setters

func (o Integration) MarshalJSON() ([]byte, error) {
	type noMethod Integration
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetGKEIntegration sets the GKE integration
func (o *Integration) SetGKE(v *GKEIntegration) *Integration {
	if o.GKE = v; o.GKE == nil {
		o.nullFields = append(o.nullFields, "GKE")
	}
	return o
}

// SetDockerSwarm sets the DockerSwarm integration
func (o *Integration) SetDockerSwarm(v *DockerSwarmIntegration) *Integration {
	if o.DockerSwarm = v; o.DockerSwarm == nil {
		o.nullFields = append(o.nullFields, "DockerSwarm")
	}
	return o
}

// region GKE integration setters

func (o GKEIntegration) MarshalJSON() ([]byte, error) {
	type noMethod GKEIntegration
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetAutoUpdate sets the autoupdate flag
func (o *GKEIntegration) SetAutoUpdate(v *bool) *GKEIntegration {
	if o.AutoUpdate = v; o.AutoUpdate == nil {
		o.nullFields = append(o.nullFields, "AutoUpdate")
	}
	return o
}

// SetAutoScale sets the AutoScale configuration used with the GKE integration
func (o *GKEIntegration) SetAutoScale(v *AutoScaleGKE) *GKEIntegration {
	if o.AutoScale = v; o.AutoScale == nil {
		o.nullFields = append(o.nullFields, "AutoScale")
	}
	return o
}

// SetLocation sets the location that the cluster is located in
func (o *GKEIntegration) SetLocation(v *string) *GKEIntegration {
	if o.Location = v; o.Location == nil {
		o.nullFields = append(o.nullFields, "Location")
	}
	return o
}

// SetClusterID sets the cluster ID
func (o *GKEIntegration) SetClusterID(v *string) *GKEIntegration {
	if o.ClusterID = v; o.ClusterID == nil {
		o.nullFields = append(o.nullFields, "ClusterID")
	}
	return o
}

// region GKE AutoScaling setters

func (o AutoScaleGKE) MarshalJSON() ([]byte, error) {
	type noMethod AutoScaleGKE
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetLabels sets the AutoScale labels for the GKE integration
func (o *AutoScaleGKE) SetLabels(v []*AutoScaleLabel) *AutoScaleGKE {
	if o.Labels = v; o.Labels == nil {
		o.nullFields = append(o.nullFields, "Labels")
	}
	return o
}

// endregion

// endregion

// region DockerSwarm integration setters

func (o DockerSwarmIntegration) MarshalJSON() ([]byte, error) {
	type noMethod DockerSwarmIntegration
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetMasterPort sets the master port
func (o *DockerSwarmIntegration) SetMasterPort(v *int) *DockerSwarmIntegration {
	if o.MasterPort = v; o.MasterPort == nil {
		o.nullFields = append(o.nullFields, "MasterPort")
	}
	return o
}

// SetMasterHost sets the master host
func (o *DockerSwarmIntegration) SetMasterHost(v *string) *DockerSwarmIntegration {
	if o.MasterHost = v; o.MasterHost == nil {
		o.nullFields = append(o.nullFields, "MasterHost")
	}
	return o
}

// endregion

// endregion

// region Scaling Policy setters

func (o Scaling) MarshalJSON() ([]byte, error) {
	type noMethod Scaling
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetUp sets the scaling policy to usewhen increasing the number of instances in a group.
func (o *Scaling) SetUp(v []*ScalingPolicy) *Scaling {
	if o.Up = v; o.Up == nil {
		o.nullFields = append(o.nullFields, "Up")
	}
	return o
}

// SetDown sets the scaling policy to use when decreasing the number of instances in a group.
func (o *Scaling) SetDown(v []*ScalingPolicy) *Scaling {
	if o.Down = v; o.Down == nil {
		o.nullFields = append(o.nullFields, "Down")
	}
	return o
}

// region ScalingPolicy setters

func (o ScalingPolicy) MarshalJSON() ([]byte, error) {
	type noMethod ScalingPolicy
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetAction sets the action to perform when scaling
func (o *ScalingPolicy) SetAction(v *Action) *ScalingPolicy {
	if o.Action = v; o.Action == nil {
		o.nullFields = append(o.nullFields, "Action")
	}
	return o
}

// SetCooldown sets the cooldown time in seconds before triggered events can start
func (o *ScalingPolicy) SetCooldown(v *int) *ScalingPolicy {
	if o.Cooldown = v; o.Cooldown == nil {
		o.nullFields = append(o.nullFields, "Cooldown")
	}
	return o
}

// SetDimensions sets the list of dimension objects
func (o *ScalingPolicy) SetDimensions(v []*Dimension) *ScalingPolicy {
	if o.Dimensions = v; o.Dimensions == nil {
		o.nullFields = append(o.nullFields, "Dimensions")
	}
	return o
}

// SetEvaluationPeriods sets the number of periods over which data is compared
func (o *ScalingPolicy) SetEvaluationPeriods(v *int) *ScalingPolicy {
	if o.EvaluationPeriods = v; o.EvaluationPeriods == nil {
		o.nullFields = append(o.nullFields, "EvaluationPeriods")
	}
	return o
}

// SetMetricName sets the name of the metric to compare
func (o *ScalingPolicy) SetMetricName(v *string) *ScalingPolicy {
	if o.MetricName = v; o.MetricName == nil {
		o.nullFields = append(o.nullFields, "MetricName")
	}
	return o
}

// SetNamespace sets the namespace for the associated metric
func (o *ScalingPolicy) SetNamespace(v *string) *ScalingPolicy {
	if o.Namespace = v; o.Namespace == nil {
		o.nullFields = append(o.nullFields, "Namespace")
	}
	return o
}

// SetOperator sets the operator (gte, lte)
func (o *ScalingPolicy) SetOperator(v *string) *ScalingPolicy {
	if o.Operator = v; o.Operator == nil {
		o.nullFields = append(o.nullFields, "Operator")
	}
	return o
}

// SetPeriod sets the period in seconds over which the statistic is applied
func (o *ScalingPolicy) SetPeriod(v *int) *ScalingPolicy {
	if o.Period = v; o.Period == nil {
		o.nullFields = append(o.nullFields, "Period")
	}
	return o
}

// SetPolicyName sets the name of the scaling policy
func (o *ScalingPolicy) SetPolicyName(v *string) *ScalingPolicy {
	if o.PolicyName = v; o.PolicyName == nil {
		o.nullFields = append(o.nullFields, "PolicyName")
	}
	return o
}

// SetSource sets the source of the metric (spectrum, stackdriver)
func (o *ScalingPolicy) SetSource(v *string) *ScalingPolicy {
	if o.Source = v; o.Source == nil {
		o.nullFields = append(o.nullFields, "Source")
	}
	return o
}

// SetStatistic sets the metric aggregator to return (average, sum, min, max)
func (o *ScalingPolicy) SetStatistic(v *string) *ScalingPolicy {
	if o.Statistic = v; o.Statistic == nil {
		o.nullFields = append(o.nullFields, "Statistic")
	}
	return o
}

// SetThreshold sets the value against which the metric is compared
func (o *ScalingPolicy) SetThreshold(v *float64) *ScalingPolicy {
	if o.Threshold = v; o.Threshold == nil {
		o.nullFields = append(o.nullFields, "Threshold")
	}
	return o
}

// SetUnit sets the unit for the associated metric
func (o *ScalingPolicy) SetUnit(v *string) *ScalingPolicy {
	if o.Unit = v; o.Unit == nil {
		o.nullFields = append(o.nullFields, "Unit")
	}
	return o
}

// region Action setters

func (o Action) MarshalJSON() ([]byte, error) {
	type noMethod Action
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetAdjustment sets the number associated with the action type
func (o *Action) SetAdjustment(v *int) *Action {
	if o.Adjustment = v; o.Adjustment == nil {
		o.nullFields = append(o.nullFields, "Adjustment")
	}
	return o
}

// SetType sets the type of action to take when scaling (adjustment)
func (o *Action) SetType(v *string) *Action {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

// endregion

// region Dimension setters

func (o Dimension) MarshalJSON() ([]byte, error) {
	type noMethod Dimension
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetName sets the name of the dimension
func (o *Dimension) SetName(v *string) *Dimension {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

// SetValue sets the value of the dimension
func (o *Dimension) SetValue(v *string) *Dimension {
	if o.Value = v; o.Value == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}

// endregion

// endregion

// endregion

// region Scheduling

func (o Scheduling) MarshalJSON() ([]byte, error) {
	type noMethod Scheduling
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Scheduling) SetTasks(v []*Task) *Scheduling {
	if o.Tasks = v; o.Tasks == nil {
		o.nullFields = append(o.nullFields, "Tasks")
	}
	return o
}

// endregion

// region Task

func (o Task) MarshalJSON() ([]byte, error) {
	type noMethod Task
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Task) SetIsEnabled(v *bool) *Task {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}

func (o *Task) SetType(v *string) *Task {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o *Task) SetCronExpression(v *string) *Task {
	if o.CronExpression = v; o.CronExpression == nil {
		o.nullFields = append(o.nullFields, "CronExpression")
	}
	return o
}

func (o *Task) SetTargetCapacity(v *int) *Task {
	if o.TargetCapacity = v; o.TargetCapacity == nil {
		o.nullFields = append(o.nullFields, "TargetCapacity")
	}
	return o
}

func (o *Task) SetMinCapacity(v *int) *Task {
	if o.MinCapacity = v; o.MinCapacity == nil {
		o.nullFields = append(o.nullFields, "MinCapacity")
	}
	return o
}

func (o *Task) SetMaxCapacity(v *int) *Task {
	if o.MaxCapacity = v; o.MaxCapacity == nil {
		o.nullFields = append(o.nullFields, "MaxCapacity")
	}
	return o
}

// endregion

// region Strategy setters

func (o Strategy) MarshalJSON() ([]byte, error) {
	type noMethod Strategy
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetDrainingTimeout sets the time to keep an instance alive after detaching it from the group
func (o *Strategy) SetDrainingTimeout(v *int) *Strategy {
	if o.DrainingTimeout = v; o.DrainingTimeout == nil {
		o.nullFields = append(o.nullFields, "DrainingTimeout")
	}
	return o
}

// SetFallbackToOnDemand sets the option to fallback to on demand instances if preemptible instances arent available
func (o *Strategy) SetFallbackToOnDemand(v *bool) *Strategy {
	if o.FallbackToOnDemand = v; o.FallbackToOnDemand == nil {
		o.nullFields = append(o.nullFields, "FallbackToOnDemand")
	}
	return o
}

// SetPreemptiblePercentage sets the ratio of preemptible instances to use in the group
func (o *Strategy) SetPreemptiblePercentage(v *int) *Strategy {
	if o.PreemptiblePercentage = v; o.PreemptiblePercentage == nil {
		o.nullFields = append(o.nullFields, "PreemptiblePercentage")
	}
	return o
}

// SetOnDemandCount sets the number of on demand instances to use in the group.
func (o *Strategy) SetOnDemandCount(v *int) *Strategy {
	if o.OnDemandCount = v; o.OnDemandCount == nil {
		o.nullFields = append(o.nullFields, "OnDemandCount")
	}
	return o
}

// endregion
