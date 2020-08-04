package aws

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
	Region              *string     `json:"region,omitempty"`
	Strategy            *Strategy   `json:"strategy,omitempty"`
	Capacity            *Capacity   `json:"capacity,omitempty"`
	Compute             *Compute    `json:"compute,omitempty"`
	Scheduling          *Scheduling `json:"scheduling,omitempty"`
	AutoScaler          *AutoScaler `json:"autoScaler,omitempty"`

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
	SpotPercentage           *float64 `json:"spotPercentage,omitempty"`
	UtilizeReservedInstances *bool    `json:"utilizeReservedInstances,omitempty"`
	FallbackToOnDemand       *bool    `json:"fallbackToOd,omitempty"`
	DrainingTimeout          *int     `json:"drainingTimeout,omitempty"`
	GracePeriod              *int     `json:"gracePeriod,omitempty"`

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
	InstanceTypes       *InstanceTypes       `json:"instanceTypes,omitempty"`
	LaunchSpecification *LaunchSpecification `json:"launchSpecification,omitempty"`
	SubnetIDs           []string             `json:"subnetIds,omitempty"`

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
	IsEnabled      *bool   `json:"isEnabled,omitempty"`
	Type           *string `json:"taskType,omitempty"`
	CronExpression *string `json:"cronExpression,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type InstanceTypes struct {
	Whitelist []string `json:"whitelist,omitempty"`
	Blacklist []string `json:"blacklist,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type LaunchSpecification struct {
	AssociatePublicIPAddress *bool               `json:"associatePublicIpAddress,omitempty"`
	SecurityGroupIDs         []string            `json:"securityGroupIds,omitempty"`
	ImageID                  *string             `json:"imageId,omitempty"`
	KeyPair                  *string             `json:"keyPair,omitempty"`
	UserData                 *string             `json:"userData,omitempty"`
	IAMInstanceProfile       *IAMInstanceProfile `json:"iamInstanceProfile,omitempty"`
	Tags                     []*Tag              `json:"tags,omitempty"`
	LoadBalancers            []*LoadBalancer     `json:"loadBalancers,omitempty"`
	RootVolumeSize           *int                `json:"rootVolumeSize,omitempty"`
	Monitoring               *bool               `json:"monitoring,omitempty"`
	EBSOptimized             *bool               `json:"ebsOptimized,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type IAMInstanceProfile struct {
	ARN  *string `json:"arn,omitempty"`
	Name *string `json:"name,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type LoadBalancer struct {
	Name *string `json:"name,omitempty"`
	Arn  *string `json:"arn,omitempty"`
	Type *string `json:"type,omitempty"`

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

type AutoScalerHeadroom struct {
	CPUPerUnit    *int `json:"cpuPerUnit,omitempty"`
	GPUPerUnit    *int `json:"gpuPerUnit,omitempty"`
	MemoryPerUnit *int `json:"memoryPerUnit,omitempty"`
	NumOfUnits    *int `json:"numOfUnits,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AutoScalerResourceLimits struct {
	MaxVCPU      *int `json:"maxVCpu,omitempty"`
	MaxMemoryGiB *int `json:"maxMemoryGib,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AutoScalerDown struct {
	EvaluationPeriods      *int     `json:"evaluationPeriods,omitempty"`
	MaxScaleDownPercentage *float64 `json:"maxScaleDownPercentage,omitempty"`

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

// Deprecated: Use CreateRollInput instead.
type RollClusterInput struct {
	Roll *Roll `json:"roll,omitempty"`
}

// Deprecated: Use CreateRollOutput instead.
type RollClusterOutput struct {
	RollClusterStatus *RollClusterStatus `json:"clusterDeploymentStatus,omitempty"`
}

// Deprecated: Use RollSpec instead.
type Roll struct {
	ClusterID                    *string  `json:"clusterId,omitempty"`
	Comment                      *string  `json:"comment,omitempty"`
	BatchSizePercentage          *int     `json:"batchSizePercentage,omitempty"`
	DisableLaunchSpecAutoscaling *bool    `json:"disableLaunchSpecAutoscaling,omitempty"`
	LaunchSpecIDs                []string `json:"launchSpecIds,omitempty"`
	InstanceIDs                  []string `json:"instanceIds,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// Deprecated: Use RollStatus instead.
type RollClusterStatus struct {
	OceanID      *string   `json:"oceanId,omitempty"`
	RollID       *string   `json:"id,omitempty"`
	RollStatus   *string   `json:"status,omitempty"`
	Progress     *Progress `json:"progress,omitempty"`
	CurrentBatch *int      `json:"currentBatch,omitempty"`
	NumOfBatches *int      `json:"numOfBatches,omitempty"`
	CreatedAt    *string   `json:"createdAt,omitempty"`
	UpdatedAt    *string   `json:"updatedAt,omitempty"`
}

type RollSpec struct {
	ID                           *string  `json:"id,omitempty"`
	ClusterID                    *string  `json:"clusterId,omitempty"`
	Comment                      *string  `json:"comment,omitempty"`
	BatchSizePercentage          *int     `json:"batchSizePercentage,omitempty"`
	DisableLaunchSpecAutoScaling *bool    `json:"disableLaunchSpecAutoScaling,omitempty"`
	LaunchSpecIDs                []string `json:"launchSpecIds,omitempty"`
	InstanceIDs                  []string `json:"instanceIds,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type RollStatus struct {
	ID            *string   `json:"id,omitempty"`
	ClusterID     *string   `json:"oceanId,omitempty"`
	Comment       *string   `json:"comment,omitempty"`
	Status        *string   `json:"status,omitempty"`
	Progress      *Progress `json:"progress,omitempty"`
	CurrentBatch  *int      `json:"currentBatch,omitempty"`
	NumOfBatches  *int      `json:"numOfBatches,omitempty"`
	LaunchSpecIDs []string  `json:"launchSpecIds,omitempty"`
	InstanceIDs   []string  `json:"instanceIds,omitempty"`
	CreatedAt     *string   `json:"createdAt,omitempty"`
	UpdatedAt     *string   `json:"updatedAt,omitempty"`
}

type Progress struct {
	Unit  *string `json:"unit,omitempty"`
	Value *int    `json:"value,omitempty"`
}

type ListRollsInput struct {
	ClusterID *string `json:"clusterId,omitempty"`
}

type ListRollsOutput struct {
	Rolls []*RollStatus `json:"rolls,omitempty"`
}

type CreateRollInput struct {
	Roll *RollSpec `json:"roll,omitempty"`
}

type CreateRollOutput struct {
	Roll *RollStatus `json:"roll,omitempty"`
}

type ReadRollInput struct {
	RollID    *string `json:"rollId,omitempty"`
	ClusterID *string `json:"clusterId,omitempty"`
}

type ReadRollOutput struct {
	Roll *RollStatus `json:"roll,omitempty"`
}

type UpdateRollInput struct {
	Roll *RollSpec `json:"roll,omitempty"`
}

type UpdateRollOutput struct {
	Roll *RollStatus `json:"roll,omitempty"`
}

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

func rollClusterStatusFromJSON(in []byte) (*RollClusterStatus, error) {
	b := new(RollClusterStatus)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func rollClusterStatusesFromJSON(in []byte) ([]*RollClusterStatus, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*RollClusterStatus, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := rollClusterStatusFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func rollClusterStatusesFromHttpResponse(resp *http.Response) ([]*RollClusterStatus, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return rollClusterStatusesFromJSON(body)
}

func rollStatusFromJSON(in []byte) (*RollStatus, error) {
	b := new(RollStatus)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func rollStatusesFromJSON(in []byte) ([]*RollStatus, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*RollStatus, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := rollStatusFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func rollStatusesFromHttpResponse(resp *http.Response) ([]*RollStatus, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return rollStatusesFromJSON(body)
}

func (s *ServiceOp) ListClusters(ctx context.Context, input *ListClustersInput) (*ListClustersOutput, error) {
	r := client.NewRequest(http.MethodGet, "/ocean/aws/k8s/cluster")
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
	r := client.NewRequest(http.MethodPost, "/ocean/aws/k8s/cluster")
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
	path, err := uritemplates.Expand("/ocean/aws/k8s/cluster/{clusterId}", uritemplates.Values{
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
	path, err := uritemplates.Expand("/ocean/aws/k8s/cluster/{clusterId}", uritemplates.Values{
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
	path, err := uritemplates.Expand("/ocean/aws/k8s/cluster/{clusterId}", uritemplates.Values{
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

func (s *ServiceOp) ListRolls(ctx context.Context, input *ListRollsInput) (*ListRollsOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/k8s/cluster/{clusterId}/roll", uritemplates.Values{
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

	v, err := rollStatusesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ListRollsOutput)
	if len(v) > 0 {
		output.Rolls = v
	}

	return output, nil
}

func (s *ServiceOp) CreateRoll(ctx context.Context, input *CreateRollInput) (*CreateRollOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/k8s/cluster/{clusterId}/roll", uritemplates.Values{
		"clusterId": spotinst.StringValue(input.Roll.ClusterID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.Roll.ClusterID = nil

	r := client.NewRequest(http.MethodPost, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	v, err := rollStatusesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateRollOutput)
	if len(v) > 0 {
		output.Roll = v[0]
	}

	return output, nil
}

func (s *ServiceOp) ReadRoll(ctx context.Context, input *ReadRollInput) (*ReadRollOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/k8s/cluster/{clusterId}/roll/{rollId}", uritemplates.Values{
		"clusterId": spotinst.StringValue(input.ClusterID),
		"rollId":    spotinst.StringValue(input.RollID),
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

	v, err := rollStatusesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadRollOutput)
	if len(v) > 0 {
		output.Roll = v[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateRoll(ctx context.Context, input *UpdateRollInput) (*UpdateRollOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/k8s/cluster/{clusterId}/roll/{rollId}", uritemplates.Values{
		"clusterId": spotinst.StringValue(input.Roll.ClusterID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.Roll.ClusterID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	v, err := rollStatusesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(UpdateRollOutput)
	if len(v) > 0 {
		output.Roll = v[0]
	}

	return output, nil
}

// Deprecated: Use CreateRoll instead.
func (s *ServiceOp) Roll(ctx context.Context, input *RollClusterInput) (*RollClusterOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/k8s/cluster/{clusterId}/roll", uritemplates.Values{
		"clusterId": spotinst.StringValue(input.Roll.ClusterID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.Roll.ClusterID = nil

	r := client.NewRequest(http.MethodPost, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rs, err := rollClusterStatusesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(RollClusterOutput)
	if len(rs) > 0 {
		output.RollClusterStatus = rs[0]
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

func (o *Cluster) SetRegion(v *string) *Cluster {
	if o.Region = v; o.Region == nil {
		o.nullFields = append(o.nullFields, "Region")
	}
	return o
}

func (o *Cluster) SetStrategy(v *Strategy) *Cluster {
	if o.Strategy = v; o.Strategy == nil {
		o.nullFields = append(o.nullFields, "Strategy")
	}
	return o
}

func (o *Cluster) SetCapacity(v *Capacity) *Cluster {
	if o.Capacity = v; o.Capacity == nil {
		o.nullFields = append(o.nullFields, "Capacity")
	}
	return o
}

func (o *Cluster) SetCompute(v *Compute) *Cluster {
	if o.Compute = v; o.Compute == nil {
		o.nullFields = append(o.nullFields, "Compute")
	}
	return o
}

func (o *Cluster) SetScheduling(v *Scheduling) *Cluster {
	if o.Scheduling = v; o.Scheduling == nil {
		o.nullFields = append(o.nullFields, "Scheduling")
	}
	return o
}

func (o *Cluster) SetAutoScaler(v *AutoScaler) *Cluster {
	if o.AutoScaler = v; o.AutoScaler == nil {
		o.nullFields = append(o.nullFields, "AutoScaler")
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

func (o *Strategy) SetSpotPercentage(v *float64) *Strategy {
	if o.SpotPercentage = v; o.SpotPercentage == nil {
		o.nullFields = append(o.nullFields, "SpotPercentage")
	}
	return o
}

func (o *Strategy) SetUtilizeReservedInstances(v *bool) *Strategy {
	if o.UtilizeReservedInstances = v; o.UtilizeReservedInstances == nil {
		o.nullFields = append(o.nullFields, "UtilizeReservedInstances")
	}
	return o
}

func (o *Strategy) SetFallbackToOnDemand(v *bool) *Strategy {
	if o.FallbackToOnDemand = v; o.FallbackToOnDemand == nil {
		o.nullFields = append(o.nullFields, "FallbackToOnDemand")
	}
	return o
}

func (o *Strategy) SetDrainingTimeout(v *int) *Strategy {
	if o.DrainingTimeout = v; o.DrainingTimeout == nil {
		o.nullFields = append(o.nullFields, "DrainingTimeout")
	}
	return o
}

func (o *Strategy) SetGracePeriod(v *int) *Strategy {
	if o.GracePeriod = v; o.GracePeriod == nil {
		o.nullFields = append(o.nullFields, "GracePeriod")
	}
	return o
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

func (o *Compute) SetLaunchSpecification(v *LaunchSpecification) *Compute {
	if o.LaunchSpecification = v; o.LaunchSpecification == nil {
		o.nullFields = append(o.nullFields, "LaunchSpecification")
	}
	return o
}

func (o *Compute) SetSubnetIDs(v []string) *Compute {
	if o.SubnetIDs = v; o.SubnetIDs == nil {
		o.nullFields = append(o.nullFields, "SubnetIDs")
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

func (o *InstanceTypes) SetBlacklist(v []string) *InstanceTypes {
	if o.Blacklist = v; o.Blacklist == nil {
		o.nullFields = append(o.nullFields, "Blacklist")
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

func (o *LaunchSpecification) SetAssociatePublicIPAddress(v *bool) *LaunchSpecification {
	if o.AssociatePublicIPAddress = v; o.AssociatePublicIPAddress == nil {
		o.nullFields = append(o.nullFields, "AssociatePublicIPAddress")
	}
	return o
}

func (o *LaunchSpecification) SetSecurityGroupIDs(v []string) *LaunchSpecification {
	if o.SecurityGroupIDs = v; o.SecurityGroupIDs == nil {
		o.nullFields = append(o.nullFields, "SecurityGroupIDs")
	}
	return o
}

func (o *LaunchSpecification) SetImageId(v *string) *LaunchSpecification {
	if o.ImageID = v; o.ImageID == nil {
		o.nullFields = append(o.nullFields, "ImageID")
	}
	return o
}

func (o *LaunchSpecification) SetKeyPair(v *string) *LaunchSpecification {
	if o.KeyPair = v; o.KeyPair == nil {
		o.nullFields = append(o.nullFields, "KeyPair")
	}
	return o
}

func (o *LaunchSpecification) SetUserData(v *string) *LaunchSpecification {
	if o.UserData = v; o.UserData == nil {
		o.nullFields = append(o.nullFields, "UserData")
	}
	return o
}

func (o *LaunchSpecification) SetIAMInstanceProfile(v *IAMInstanceProfile) *LaunchSpecification {
	if o.IAMInstanceProfile = v; o.IAMInstanceProfile == nil {
		o.nullFields = append(o.nullFields, "IAMInstanceProfile")
	}
	return o
}

func (o *LaunchSpecification) SetTags(v []*Tag) *LaunchSpecification {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

func (o *LaunchSpecification) SetLoadBalancers(v []*LoadBalancer) *LaunchSpecification {
	if o.LoadBalancers = v; o.LoadBalancers == nil {
		o.nullFields = append(o.nullFields, "LoadBalancers")
	}
	return o
}

func (o *LaunchSpecification) SetRootVolumeSize(v *int) *LaunchSpecification {
	if o.RootVolumeSize = v; o.RootVolumeSize == nil {
		o.nullFields = append(o.nullFields, "RootVolumeSize")
	}
	return o
}

func (o *LaunchSpecification) SetMonitoring(v *bool) *LaunchSpecification {
	if o.Monitoring = v; o.Monitoring == nil {
		o.nullFields = append(o.nullFields, "Monitoring")
	}
	return o
}

func (o *LaunchSpecification) SetEBSOptimized(v *bool) *LaunchSpecification {
	if o.EBSOptimized = v; o.EBSOptimized == nil {
		o.nullFields = append(o.nullFields, "EBSOptimized")
	}
	return o
}

// endregion

// region LoadBalancer

func (o *LoadBalancer) SetArn(v *string) *LoadBalancer {
	if o.Arn = v; o.Arn == nil {
		o.nullFields = append(o.nullFields, "Arn")
	}
	return o
}

func (o *LoadBalancer) SetName(v *string) *LoadBalancer {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *LoadBalancer) SetType(v *string) *LoadBalancer {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

// endregion

// region IAMInstanceProfile

func (o IAMInstanceProfile) MarshalJSON() ([]byte, error) {
	type noMethod IAMInstanceProfile
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *IAMInstanceProfile) SetArn(v *string) *IAMInstanceProfile {
	if o.ARN = v; o.ARN == nil {
		o.nullFields = append(o.nullFields, "ARN")
	}
	return o
}

func (o *IAMInstanceProfile) SetName(v *string) *IAMInstanceProfile {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
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

func (o *AutoScalerHeadroom) SetGPUPerUnit(v *int) *AutoScalerHeadroom {
	if o.GPUPerUnit = v; o.GPUPerUnit == nil {
		o.nullFields = append(o.nullFields, "GPUPerUnit")
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

// region Roll

func (o Roll) MarshalJSON() ([]byte, error) {
	type noMethod Roll
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Roll) SetComment(v *string) *Roll {
	if o.Comment = v; o.Comment == nil {
		o.nullFields = append(o.nullFields, "Comment")
	}
	return o
}

func (o *Roll) SetBatchSizePercentage(v *int) *Roll {
	if o.BatchSizePercentage = v; o.BatchSizePercentage == nil {
		o.nullFields = append(o.nullFields, "BatchSizePercentage")
	}
	return o
}

func (o *Roll) SetDisableLaunchSpecAutoscaling(v *bool) *Roll {
	if o.DisableLaunchSpecAutoscaling = v; o.DisableLaunchSpecAutoscaling == nil {
		o.nullFields = append(o.nullFields, "DisableLaunchSpecAutoscaling")
	}
	return o
}

func (o *Roll) SetLaunchSpecIDs(v []string) *Roll {
	if o.LaunchSpecIDs = v; o.LaunchSpecIDs == nil {
		o.nullFields = append(o.nullFields, "LaunchSpecIDs")
	}
	return o
}

func (o *Roll) SetInstanceIDs(v []string) *Roll {
	if o.InstanceIDs = v; o.InstanceIDs == nil {
		o.nullFields = append(o.nullFields, "InstanceIDs")
	}
	return o
}

// endregion
