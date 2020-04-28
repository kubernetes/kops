package gcp

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/jsonutil"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/uritemplates"
)

type Cluster struct {
	ID                  *string     `json:"id,omitempty"`
	ControllerClusterID *string     `json:"controllerClusterId,omitempty"`
	Name                *string     `json:"name,omitempty"`
	Scheduling          *Scheduling `json:"scheduling,omitempty"`
	AutoScaler          *AutoScaler `json:"autoScaler,omitempty"`
	Capacity            *Capacity   `json:"capacity,omitempty"`
	Compute             *Compute    `json:"compute,omitempty"`
	Strategy            *Strategy   `json:"strategy,omitempty"`
	GKE                 *GKE        `json:"gke,omitempty"`

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

type Strategy struct {
	DrainingTimeout *int `json:"drainingTimeout,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AutoScaler struct {
	IsEnabled              *bool                     `json:"isEnabled,omitempty"`
	IsAutoConfig           *bool                     `json:"isAutoConfig,omitempty"`
	Cooldown               *int                      `json:"cooldown,omitempty"`
	AutoHeadroomPercentage *int                      `json:"autoHeadroomPercentage,omitempty"`
	Headroom               *AutoScalerHeadroom       `json:"headroom,omitempty"`
	ResourceLimits         *AutoScalerResourceLimits `json:"resourceLimits,omitempty"`
	Down                   *AutoScalerDown           `json:"down,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AutoScalerDown struct {
	EvaluationPeriods      *int     `json:"evaluationPeriods,omitempty"`
	MaxScaleDownPercentage *float64 `json:"maxScaleDownPercentage,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AutoScalerHeadroom struct {
	CPUPerUnit    *int `json:"cpuPerUnit,omitempty"`
	MemoryPerUnit *int `json:"memoryPerUnit,omitempty"`
	NumOfUnits    *int `json:"numOfUnits,omitempty"`
	GPUPerUnit    *int `json:"gpuPerUnit,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AutoScalerResourceLimits struct {
	MaxVCPU      *int `json:"maxVCpu,omitempty"`
	MaxMemoryGiB *int `json:"maxMemoryGib,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type BackendService struct {
	BackendServiceName *string     `json:"backendServiceName,omitempty"`
	LocationType       *string     `json:"locationType,omitempty"`
	Scheme             *string     `json:"scheme,omitempty"`
	NamedPorts         *NamedPorts `json:"namedPorts,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Capacity struct {
	Minimum *int `json:"minimum,omitempty"`
	Maximum *int `json:"maximum,omitempty"`
	Target  *int `json:"target,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Compute struct {
	AvailabilityZones   []string             `json:"availabilityZones,omitempty"`
	InstanceTypes       *InstanceTypes       `json:"instanceTypes,omitempty"`
	LaunchSpecification *LaunchSpecification `json:"launchSpecification,omitempty"`
	BackendServices     []*BackendService    `json:"backendServices,omitempty"`
	NetworkInterfaces   []*NetworkInterface  `json:"networkInterfaces,omitempty"`
	SubnetName          *string              `json:"subnetName,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Scheduling struct {
	ShutdownHours *ShutdownHours `json:"shutdownHours,omitempty"`
	Tasks         []*Task        `json:"tasks,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ShutdownHours struct {
	IsEnabled   *bool    `json:"isEnabled,omitempty"`
	TimeWindows []string `json:"timeWindows,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Task struct {
	IsEnabled           *bool   `json:"isEnabled,omitempty"`
	Type                *string `json:"taskType,omitempty"`
	CronExpression      *string `json:"cronExpression,omitempty"`
	BatchSizePercentage *int    `json:"batchSizePercentage,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type GKE struct {
	ClusterName    *string `json:"clusterName,omitempty"`
	MasterLocation *string `json:"masterLocation,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type InstanceTypes struct {
	Whitelist []string `json:"whitelist,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type LaunchSpecification struct {
	Labels             []*Label    `json:"labels,omitempty"`
	IPForwarding       *bool       `json:"ipForwarding,omitempty"`
	Metadata           []*Metadata `json:"metadata,omitempty"`
	RootVolumeSizeInGB *int        `json:"rootVolumeSizeInGb,omitempty"`
	ServiceAccount     *string     `json:"serviceAccount,omitempty"`
	SourceImage        *string     `json:"sourceImage,omitempty"`
	Tags               []string    `json:"tags,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Metadata struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type NamedPorts struct {
	Name  *string `json:"name,omitempty"`
	Ports []int   `json:"ports,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type NetworkInterface struct {
	AccessConfigs []*AccessConfig `json:"accessConfigs,omitempty"`
	AliasIPRanges []*AliasIPRange `json:"aliasIpRanges,omitempty"`
	Network       *string         `json:"network,omitempty"`
	ProjectID     *string         `json:"projectId,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AccessConfig struct {
	Name *string `json:"name,omitempty"`
	Type *string `json:"type,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AliasIPRange struct {
	IPCIDRRange         *string `json:"ipCidrRange,omitempty"`
	SubnetworkRangeName *string `json:"subnetworkRangeName,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ListClustersInput struct{}

type ListClustersOutput struct {
	Clusters []*Cluster `json:"clusters,omitempty"`
}

type CreateClusterInput struct {
	Cluster *Cluster `json:"cluster,omitempty"`
}

type CreateClusterOutput struct {
	Cluster *Cluster `json:"cluster,omitempty"`
}

type ReadClusterInput struct {
	ClusterID *string `json:"clusterId,omitempty"`
}

type ReadClusterOutput struct {
	Cluster *Cluster `json:"cluster,omitempty"`
}

type UpdateClusterInput struct {
	Cluster *Cluster `json:"cluster,omitempty"`
}

type UpdateClusterOutput struct {
	Cluster *Cluster `json:"cluster,omitempty"`
}

type DeleteClusterInput struct {
	ClusterID *string `json:"clusterId,omitempty"`
}

type DeleteClusterOutput struct{}

func clusterFromJSON(in []byte) (*Cluster, error) {
	b := new(Cluster)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func clustersFromJSON(in []byte) ([]*Cluster, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Cluster, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := clusterFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func clustersFromHttpResponse(resp *http.Response) ([]*Cluster, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return clustersFromJSON(body)
}

func clusterImportFromJSON(in []byte) (*ImportOceanGKEClusterOutput, error) {
	b := new(ImportOceanGKEClusterOutput)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func clustersImportFromJSON(in []byte) ([]*ImportOceanGKEClusterOutput, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*ImportOceanGKEClusterOutput, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := clusterImportFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func clustersImportFromHttpResponse(resp *http.Response) ([]*ImportOceanGKEClusterOutput, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return clustersImportFromJSON(body)
}

func (s *ServiceOp) ListClusters(ctx context.Context, input *ListClustersInput) (*ListClustersOutput, error) {
	r := client.NewRequest(http.MethodGet, "/ocean/gcp/k8s/cluster")
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := clustersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListClustersOutput{Clusters: gs}, nil
}

func (s *ServiceOp) CreateCluster(ctx context.Context, input *CreateClusterInput) (*CreateClusterOutput, error) {
	r := client.NewRequest(http.MethodPost, "/ocean/gcp/k8s/cluster")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := clustersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateClusterOutput)
	if len(gs) > 0 {
		output.Cluster = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) ReadCluster(ctx context.Context, input *ReadClusterInput) (*ReadClusterOutput, error) {
	path, err := uritemplates.Expand("/ocean/gcp/k8s/cluster/{clusterId}", uritemplates.Values{
		"clusterId": spotinst.StringValue(input.ClusterID),
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

	gs, err := clustersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadClusterOutput)
	if len(gs) > 0 {
		output.Cluster = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateCluster(ctx context.Context, input *UpdateClusterInput) (*UpdateClusterOutput, error) {
	path, err := uritemplates.Expand("/ocean/gcp/k8s/cluster/{clusterId}", uritemplates.Values{
		"clusterId": spotinst.StringValue(input.Cluster.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do NOT need the ID anymore, so let's drop it.
	input.Cluster.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := clustersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(UpdateClusterOutput)
	if len(gs) > 0 {
		output.Cluster = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) DeleteCluster(ctx context.Context, input *DeleteClusterInput) (*DeleteClusterOutput, error) {
	path, err := uritemplates.Expand("/ocean/gcp/k8s/cluster/{clusterId}", uritemplates.Values{
		"clusterId": spotinst.StringValue(input.ClusterID),
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

	return &DeleteClusterOutput{}, nil
}

// ImportOceanGKECluster imports an existing Ocean GKE cluster into Elastigroup.
func (s *ServiceOp) ImportOceanGKECluster(ctx context.Context, input *ImportOceanGKEClusterInput) (*ImportOceanGKEClusterOutput, error) {
	r := client.NewRequest(http.MethodPost, "/ocean/gcp/k8s/cluster/gke/import")

	r.Params["location"] = []string{spotinst.StringValue(input.Location)}
	r.Params["clusterName"] = []string{spotinst.StringValue(input.ClusterName)}

	body := &ImportOceanGKEClusterInput{Cluster: input.Cluster}
	r.Obj = body

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := clustersImportFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ImportOceanGKEClusterOutput)
	if len(gs) > 0 {
		output = gs[0]
	}

	return output, nil
}

// region Cluster

func (o Cluster) MarshalJSON() ([]byte, error) {
	type noMethod Cluster
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Cluster) SetId(v *string) *Cluster {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *Cluster) SetControllerClusterId(v *string) *Cluster {
	if o.ControllerClusterID = v; o.ControllerClusterID == nil {
		o.nullFields = append(o.nullFields, "ControllerClusterID")
	}
	return o
}

func (o *Cluster) SetName(v *string) *Cluster {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *Cluster) SetCapacity(v *Capacity) *Cluster {
	if o.Capacity = v; o.Capacity == nil {
		o.nullFields = append(o.nullFields, "Capacity")
	}
	return o
}

func (o *Cluster) SetStrategy(v *Strategy) *Cluster {
	if o.Strategy = v; o.Strategy == nil {
		o.nullFields = append(o.nullFields, "Strategy")
	}
	return o
}

func (o *Cluster) SetCompute(v *Compute) *Cluster {
	if o.Compute = v; o.Compute == nil {
		o.nullFields = append(o.nullFields, "Compute")
	}
	return o
}

func (o *Cluster) SetAutoScaler(v *AutoScaler) *Cluster {
	if o.AutoScaler = v; o.AutoScaler == nil {
		o.nullFields = append(o.nullFields, "AutoScaler")
	}
	return o
}

func (o *Cluster) SetGKE(v *GKE) *Cluster {
	if o.GKE = v; o.GKE == nil {
		o.nullFields = append(o.nullFields, "GKE")
	}
	return o
}

func (o *Cluster) SetScheduling(v *Scheduling) *Cluster {
	if o.Scheduling = v; o.Scheduling == nil {
		o.nullFields = append(o.nullFields, "Scheduling")
	}
	return o
}

// endregion

// region GKE

func (o GKE) MarshalJSON() ([]byte, error) {
	type noMethod GKE
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *GKE) SetClusterName(v *string) *GKE {
	if o.ClusterName = v; o.ClusterName == nil {
		o.nullFields = append(o.nullFields, "ClusterName")
	}
	return o
}

func (o *GKE) SetMasterLocation(v *string) *GKE {
	if o.MasterLocation = v; o.MasterLocation == nil {
		o.nullFields = append(o.nullFields, "MasterLocation")
	}
	return o
}

// endregion

// region Import

type ImportOceanGKECluster struct {
	InstanceTypes   *InstanceTypes    `json:"instanceTypes,omitempty"`
	BackendServices []*BackendService `json:"backendServices,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ImportOceanGKEClusterInput struct {
	ClusterName        *string                `json:"clusterName,omitempty"`
	Location           *string                `json:"location,omitempty"`
	NodePoolName       *string                `json:"nodePoolName,omitempty"`
	IncludeLaunchSpecs *string                `json:"includeLaunchSpecs,omitempty"`
	Cluster            *ImportOceanGKECluster `json:"cluster,omitempty"`
}

type ImportOceanGKEClusterOutput struct {
	Cluster *Cluster `json:"cluster,omitempty"`
}

// endregion

// region Capacity

func (o Capacity) MarshalJSON() ([]byte, error) {
	type noMethod Capacity
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Capacity) SetMinimum(v *int) *Capacity {
	if o.Minimum = v; o.Minimum == nil {
		o.nullFields = append(o.nullFields, "Minimum")
	}
	return o
}

func (o *Capacity) SetMaximum(v *int) *Capacity {
	if o.Maximum = v; o.Maximum == nil {
		o.nullFields = append(o.nullFields, "Maximum")
	}
	return o
}

func (o *Capacity) SetTarget(v *int) *Capacity {
	if o.Target = v; o.Target == nil {
		o.nullFields = append(o.nullFields, "Target")
	}
	return o
}

// endregion

// region Scheduling

func (o Scheduling) MarshalJSON() ([]byte, error) {
	type noMethod Scheduling
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Scheduling) SetShutdownHours(v *ShutdownHours) *Scheduling {
	if o.ShutdownHours = v; o.ShutdownHours == nil {
		o.nullFields = append(o.nullFields, "ShutdownHours")
	}
	return o
}

func (o *Scheduling) SetTasks(v []*Task) *Scheduling {
	if o.Tasks = v; o.Tasks == nil {
		o.nullFields = append(o.nullFields, "Tasks")
	}
	return o
}

// endregion

// region Tasks

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

func (o *Task) SetBatchSizePercentage(v *int) *Task {
	if o.BatchSizePercentage = v; o.BatchSizePercentage == nil {
		o.nullFields = append(o.nullFields, "BatchSizePercentage")
	}
	return o
}

// endregion

// region ShutdownHours

func (o ShutdownHours) MarshalJSON() ([]byte, error) {
	type noMethod ShutdownHours
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ShutdownHours) SetIsEnabled(v *bool) *ShutdownHours {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}

func (o *ShutdownHours) SetTimeWindows(v []string) *ShutdownHours {
	if o.TimeWindows = v; o.TimeWindows == nil {
		o.nullFields = append(o.nullFields, "TimeWindows")
	}
	return o
}

// endregion

// region Strategy

func (o Strategy) MarshalJSON() ([]byte, error) {
	type noMethod Strategy
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Strategy) SetDrainingTimeout(v *int) *Strategy {
	if o.DrainingTimeout = v; o.DrainingTimeout == nil {
		o.nullFields = append(o.nullFields, "DrainingTimeout")
	}
	return o
}

// endregion

// region Compute

func (o Compute) MarshalJSON() ([]byte, error) {
	type noMethod Compute
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Compute) SetInstanceTypes(v *InstanceTypes) *Compute {
	if o.InstanceTypes = v; o.InstanceTypes == nil {
		o.nullFields = append(o.nullFields, "InstanceTypes")
	}
	return o
}

func (o *Compute) SetAvailabilityZones(v []string) *Compute {
	if o.AvailabilityZones = v; o.AvailabilityZones == nil {
		o.nullFields = append(o.nullFields, "AvailabilityZones")
	}
	return o
}

func (o *Compute) SetLaunchSpecification(v *LaunchSpecification) *Compute {
	if o.LaunchSpecification = v; o.LaunchSpecification == nil {
		o.nullFields = append(o.nullFields, "LaunchSpecification")
	}
	return o
}

func (o *Compute) SetNetworkInterfaces(v []*NetworkInterface) *Compute {
	if o.NetworkInterfaces = v; o.NetworkInterfaces == nil {
		o.nullFields = append(o.nullFields, "NetworkInterfaces")
	}
	return o
}

func (o *Compute) SetSubnetName(v *string) *Compute {
	if o.SubnetName = v; o.SubnetName == nil {
		o.nullFields = append(o.nullFields, "SubnetName")
	}
	return o
}

// endregion

// region InstanceTypes

func (o InstanceTypes) MarshalJSON() ([]byte, error) {
	type noMethod InstanceTypes
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *InstanceTypes) SetWhitelist(v []string) *InstanceTypes {
	if o.Whitelist = v; o.Whitelist == nil {
		o.nullFields = append(o.nullFields, "Whitelist")
	}
	return o
}

// endregion

// region LaunchSpecification

func (o LaunchSpecification) MarshalJSON() ([]byte, error) {
	type noMethod LaunchSpecification
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (c *Compute) SetBackendServices(v []*BackendService) *Compute {
	if c.BackendServices = v; c.BackendServices == nil {
		c.nullFields = append(c.nullFields, "BackendServices")
	}
	return c
}

func (o *LaunchSpecification) SetLabels(v []*Label) *LaunchSpecification {
	if o.Labels = v; o.Labels == nil {
		o.nullFields = append(o.nullFields, "Labels")
	}
	return o
}

func (o *LaunchSpecification) SetIPForwarding(v *bool) *LaunchSpecification {
	if o.IPForwarding = v; o.IPForwarding == nil {
		o.nullFields = append(o.nullFields, "IPForwarding")
	}
	return o
}

func (o *LaunchSpecification) SetMetadata(v []*Metadata) *LaunchSpecification {
	if o.Metadata = v; o.Metadata == nil {
		o.nullFields = append(o.nullFields, "Metadata")
	}
	return o
}

func (o *LaunchSpecification) SetRootVolumeSizeInGB(v *int) *LaunchSpecification {
	if o.RootVolumeSizeInGB = v; o.RootVolumeSizeInGB == nil {
		o.nullFields = append(o.nullFields, "RootVolumeSizeInGB")
	}
	return o
}

func (o *LaunchSpecification) SetServiceAccount(v *string) *LaunchSpecification {
	if o.ServiceAccount = v; o.ServiceAccount == nil {
		o.nullFields = append(o.nullFields, "ServiceAccount")
	}
	return o
}

func (o *LaunchSpecification) SetSourceImage(v *string) *LaunchSpecification {
	if o.SourceImage = v; o.SourceImage == nil {
		o.nullFields = append(o.nullFields, "SourceImage")
	}
	return o
}

func (o *LaunchSpecification) SetTags(v []string) *LaunchSpecification {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

// endregion

// region BackendService

func (o BackendService) MarshalJSON() ([]byte, error) {
	type noMethod BackendService
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *BackendService) SetBackendServiceName(v *string) *BackendService {
	if o.BackendServiceName = v; o.BackendServiceName == nil {
		o.nullFields = append(o.nullFields, "BackendServiceName")
	}
	return o
}

func (o *BackendService) SetLocationType(v *string) *BackendService {
	if o.LocationType = v; o.LocationType == nil {
		o.nullFields = append(o.nullFields, "LocationType")
	}
	return o
}

func (o *BackendService) SetScheme(v *string) *BackendService {
	if o.Scheme = v; o.Scheme == nil {
		o.nullFields = append(o.nullFields, "Scheme")
	}
	return o
}

func (o *BackendService) SetNamedPorts(v *NamedPorts) *BackendService {
	if o.NamedPorts = v; o.NamedPorts == nil {
		o.nullFields = append(o.nullFields, "NamedPort")
	}
	return o
}

// endregion

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

func (o *NamedPorts) SetPorts(v []int) *NamedPorts {
	if o.Ports = v; o.Ports == nil {
		o.nullFields = append(o.nullFields, "Ports")
	}
	return o
}

// endregion

// region Metadata

func (o Metadata) MarshalJSON() ([]byte, error) {
	type noMethod Metadata
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Metadata) SetKey(v *string) *Metadata {
	if o.Key = v; o.Key == nil {
		o.nullFields = append(o.nullFields, "Key")
	}
	return o
}

func (o *Metadata) SetValue(v *string) *Metadata {
	if o.Value = v; o.Value == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}

// endregion

// region NetworkInterface

func (o NetworkInterface) MarshalJSON() ([]byte, error) {
	type noMethod NetworkInterface
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *NetworkInterface) SetAccessConfigs(v []*AccessConfig) *NetworkInterface {
	if o.AccessConfigs = v; o.AccessConfigs == nil {
		o.nullFields = append(o.nullFields, "AccessConfigs")
	}
	return o
}

func (o *NetworkInterface) SetAliasIPRanges(v []*AliasIPRange) *NetworkInterface {
	if o.AliasIPRanges = v; o.AliasIPRanges == nil {
		o.nullFields = append(o.nullFields, "AliasIPRanges")
	}
	return o
}

func (o *NetworkInterface) SetNetwork(v *string) *NetworkInterface {
	if o.Network = v; o.Network == nil {
		o.nullFields = append(o.nullFields, "Network")
	}
	return o
}

func (o *NetworkInterface) SetProjectId(v *string) *NetworkInterface {
	if o.ProjectID = v; o.ProjectID == nil {
		o.nullFields = append(o.nullFields, "ProjectID")
	}
	return o
}

// endregion

// region AliasIPRange

func (o AliasIPRange) MarshalJSON() ([]byte, error) {
	type noMethod AliasIPRange
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *AliasIPRange) SetIPCIDRRange(v *string) *AliasIPRange {
	if o.IPCIDRRange = v; o.IPCIDRRange == nil {
		o.nullFields = append(o.nullFields, "IPCIDRRange")
	}
	return o
}

func (o *AliasIPRange) SetSubnetworkRangeName(v *string) *AliasIPRange {
	if o.SubnetworkRangeName = v; o.SubnetworkRangeName == nil {
		o.nullFields = append(o.nullFields, "SubnetworkRangeName")
	}
	return o
}

// endregion

// region AccessConfig

func (o AccessConfig) MarshalJSON() ([]byte, error) {
	type noMethod AccessConfig
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *AccessConfig) SetName(v *string) *AccessConfig {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *AccessConfig) SetType(v *string) *AccessConfig {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

// endregion

// region AutoScaler

func (o AutoScaler) MarshalJSON() ([]byte, error) {
	type noMethod AutoScaler
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *AutoScaler) SetIsEnabled(v *bool) *AutoScaler {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}

func (o *AutoScaler) SetIsAutoConfig(v *bool) *AutoScaler {
	if o.IsAutoConfig = v; o.IsAutoConfig == nil {
		o.nullFields = append(o.nullFields, "IsAutoConfig")
	}
	return o
}

func (o *AutoScaler) SetCooldown(v *int) *AutoScaler {
	if o.Cooldown = v; o.Cooldown == nil {
		o.nullFields = append(o.nullFields, "Cooldown")
	}
	return o
}

func (o *AutoScaler) SetAutoHeadroomPercentage(v *int) *AutoScaler {
	if o.AutoHeadroomPercentage = v; o.AutoHeadroomPercentage == nil {
		o.nullFields = append(o.nullFields, "AutoHeadroomPercentage")
	}
	return o
}

func (o *AutoScaler) SetHeadroom(v *AutoScalerHeadroom) *AutoScaler {
	if o.Headroom = v; o.Headroom == nil {
		o.nullFields = append(o.nullFields, "Headroom")
	}
	return o
}

func (o *AutoScaler) SetResourceLimits(v *AutoScalerResourceLimits) *AutoScaler {
	if o.ResourceLimits = v; o.ResourceLimits == nil {
		o.nullFields = append(o.nullFields, "ResourceLimits")
	}
	return o
}

func (o *AutoScaler) SetDown(v *AutoScalerDown) *AutoScaler {
	if o.Down = v; o.Down == nil {
		o.nullFields = append(o.nullFields, "Down")
	}
	return o
}

// endregion

// region AutoScalerHeadroom

func (o AutoScalerHeadroom) MarshalJSON() ([]byte, error) {
	type noMethod AutoScalerHeadroom
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *AutoScalerHeadroom) SetCPUPerUnit(v *int) *AutoScalerHeadroom {
	if o.CPUPerUnit = v; o.CPUPerUnit == nil {
		o.nullFields = append(o.nullFields, "CPUPerUnit")
	}
	return o
}

func (o *AutoScalerHeadroom) SetMemoryPerUnit(v *int) *AutoScalerHeadroom {
	if o.MemoryPerUnit = v; o.MemoryPerUnit == nil {
		o.nullFields = append(o.nullFields, "MemoryPerUnit")
	}
	return o
}

func (o *AutoScalerHeadroom) SetNumOfUnits(v *int) *AutoScalerHeadroom {
	if o.NumOfUnits = v; o.NumOfUnits == nil {
		o.nullFields = append(o.nullFields, "NumOfUnits")
	}
	return o
}

// SetGPUPerUnit sets the gpu per unit
func (o *AutoScalerHeadroom) SetGPUPerUnit(v *int) *AutoScalerHeadroom {
	if o.GPUPerUnit = v; o.GPUPerUnit == nil {
		o.nullFields = append(o.nullFields, "GPUPerUnit")
	}
	return o
}

// endregion

// region AutoScalerResourceLimits

func (o AutoScalerResourceLimits) MarshalJSON() ([]byte, error) {
	type noMethod AutoScalerResourceLimits
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *AutoScalerResourceLimits) SetMaxVCPU(v *int) *AutoScalerResourceLimits {
	if o.MaxVCPU = v; o.MaxVCPU == nil {
		o.nullFields = append(o.nullFields, "MaxVCPU")
	}
	return o
}

func (o *AutoScalerResourceLimits) SetMaxMemoryGiB(v *int) *AutoScalerResourceLimits {
	if o.MaxMemoryGiB = v; o.MaxMemoryGiB == nil {
		o.nullFields = append(o.nullFields, "MaxMemoryGiB")
	}
	return o
}

// endregion

// region AutoScalerDown

func (o AutoScalerDown) MarshalJSON() ([]byte, error) {
	type noMethod AutoScalerDown
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *AutoScalerDown) SetEvaluationPeriods(v *int) *AutoScalerDown {
	if o.EvaluationPeriods = v; o.EvaluationPeriods == nil {
		o.nullFields = append(o.nullFields, "EvaluationPeriods")
	}
	return o
}

func (o *AutoScalerDown) SetMaxScaleDownPercentage(v *float64) *AutoScalerDown {
	if o.MaxScaleDownPercentage = v; o.MaxScaleDownPercentage == nil {
		o.nullFields = append(o.nullFields, "MaxScaleDownPercentage")
	}
	return o
}

// endregion
