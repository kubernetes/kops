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

type ECSCluster struct {
	ID          *string        `json:"id,omitempty"`
	Name        *string        `json:"name,omitempty"`
	ClusterName *string        `json:"clusterName,omitempty"`
	Region      *string        `json:"region,omitempty"`
	Capacity    *ECSCapacity   `json:"capacity,omitempty"`
	Compute     *ECSCompute    `json:"compute,omitempty"`
	AutoScaler  *ECSAutoScaler `json:"autoScaler,omitempty"`
	Strategy    *ECSStrategy   `json:"strategy,omitempty"`
	Scheduling  *ECSScheduling `json:"scheduling,omitempty"`

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

type ECSStrategy struct {
	DrainingTimeout          *int  `json:"drainingTimeout,omitempty"`
	UtilizeReservedInstances *bool `json:"utilizeReservedInstances,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSScheduling struct {
	Tasks         []*ECSTask        `json:"tasks,omitempty"`
	ShutdownHours *ECSShutdownHours `json:"shutdownHours,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSShutdownHours struct {
	IsEnabled   *bool    `json:"isEnabled,omitempty"`
	TimeWindows []string `json:"timeWindows,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSTask struct {
	IsEnabled      *bool   `json:"isEnabled,omitempty"`
	Type           *string `json:"taskType,omitempty"`
	CronExpression *string `json:"cronExpression,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSCapacity struct {
	Minimum *int `json:"minimum,omitempty"`
	Maximum *int `json:"maximum,omitempty"`
	Target  *int `json:"target,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSCompute struct {
	InstanceTypes       *ECSInstanceTypes       `json:"instanceTypes,omitempty"`
	LaunchSpecification *ECSLaunchSpecification `json:"launchSpecification,omitempty"`
	SubnetIDs           []string                `json:"subnetIds,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSInstanceTypes struct {
	Whitelist []string `json:"whitelist,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSLaunchSpecification struct {
	AssociatePublicIPAddress *bool                  `json:"associatePublicIpAddress,omitempty"`
	SecurityGroupIDs         []string               `json:"securityGroupIds,omitempty"`
	ImageID                  *string                `json:"imageId,omitempty"`
	KeyPair                  *string                `json:"keyPair,omitempty"`
	UserData                 *string                `json:"userData,omitempty"`
	IAMInstanceProfile       *ECSIAMInstanceProfile `json:"iamInstanceProfile,omitempty"`
	Tags                     []*Tag                 `json:"tags,omitempty"`
	Monitoring               *bool                  `json:"monitoring,omitempty"`
	EBSOptimized             *bool                  `json:"ebsOptimized,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSIAMInstanceProfile struct {
	ARN  *string `json:"arn,omitempty"`
	Name *string `json:"name,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSAutoScaler struct {
	IsEnabled      *bool                        `json:"isEnabled,omitempty"`
	IsAutoConfig   *bool                        `json:"isAutoConfig,omitempty"`
	Cooldown       *int                         `json:"cooldown,omitempty"`
	Headroom       *ECSAutoScalerHeadroom       `json:"headroom,omitempty"`
	ResourceLimits *ECSAutoScalerResourceLimits `json:"resourceLimits,omitempty"`
	Down           *ECSAutoScalerDown           `json:"down,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSAutoScalerHeadroom struct {
	CPUPerUnit    *int `json:"cpuPerUnit,omitempty"`
	MemoryPerUnit *int `json:"memoryPerUnit,omitempty"`
	NumOfUnits    *int `json:"numOfUnits,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSAutoScalerResourceLimits struct {
	MaxVCPU      *int `json:"maxVCpu,omitempty"`
	MaxMemoryGiB *int `json:"maxMemoryGib,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSAutoScalerDown struct {
	MaxScaleDownPercentage *int `json:"maxScaleDownPercentage,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ListECSClustersInput struct{}

type ListECSClustersOutput struct {
	Clusters []*ECSCluster `json:"clusters,omitempty"`
}

type CreateECSClusterInput struct {
	Cluster *ECSCluster `json:"cluster,omitempty"`
}

type CreateECSClusterOutput struct {
	Cluster *ECSCluster `json:"cluster,omitempty"`
}

type ReadECSClusterInput struct {
	ClusterID *string `json:"clusterId,omitempty"`
}

type ReadECSClusterOutput struct {
	Cluster *ECSCluster `json:"cluster,omitempty"`
}

type UpdateECSClusterInput struct {
	Cluster *ECSCluster `json:"cluster,omitempty"`
}

type UpdateECSClusterOutput struct {
	Cluster *ECSCluster `json:"cluster,omitempty"`
}

type DeleteECSClusterInput struct {
	ClusterID *string `json:"clusterId,omitempty"`
}

type DeleteECSClusterOutput struct{}

type ECSRollClusterInput struct {
	Roll *ECSRoll `json:"roll,omitempty"`
}

type ECSRollClusterOutput struct {
	RollClusterStatus *ECSRollClusterStatus `json:"clusterDeploymentStatus,omitempty"`
}

type ECSRoll struct {
	ClusterID           *string `json:"clusterId,omitempty"`
	BatchSizePercentage *int    `json:"batchSizePercentage,omitempty"`
}

type ECSRollClusterStatus struct {
	OceanID      *string      `json:"oceanId,omitempty"`
	RollID       *string      `json:"id,omitempty"`
	RollStatus   *string      `json:"status,omitempty"`
	Progress     *ECSProgress `json:"progress,omitempty"`
	CurrentBatch *int         `json:"currentBatch,omitempty"`
	NumOfBatches *int         `json:"numOfBatches,omitempty"`
	CreatedAt    *string      `json:"createdAt,omitempty"`
	UpdatedAt    *string      `json:"updatedAt,omitempty"`
}

type ECSProgress struct {
	Unit  *string `json:"unit,omitempty"`
	Value *int    `json:"value,omitempty"`
}

func ecsClusterFromJSON(in []byte) (*ECSCluster, error) {
	b := new(ECSCluster)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func ecsClustersFromJSON(in []byte) ([]*ECSCluster, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*ECSCluster, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := ecsClusterFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func ecsClustersFromHttpResponse(resp *http.Response) ([]*ECSCluster, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ecsClustersFromJSON(body)
}

func ecsRollStatusFromJSON(in []byte) (*ECSRollClusterStatus, error) {
	b := new(ECSRollClusterStatus)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func ecsRollStatusesFromJSON(in []byte) ([]*ECSRollClusterStatus, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*ECSRollClusterStatus, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := ecsRollStatusFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func ecsRollStatusesFromHttpResponse(resp *http.Response) ([]*ECSRollClusterStatus, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ecsRollStatusesFromJSON(body)
}

func (s *ServiceOp) ListECSClusters(ctx context.Context, input *ListECSClustersInput) (*ListECSClustersOutput, error) {
	r := client.NewRequest(http.MethodGet, "/ocean/aws/ecs/cluster")
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := ecsClustersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListECSClustersOutput{Clusters: gs}, nil
}

func (s *ServiceOp) CreateECSCluster(ctx context.Context, input *CreateECSClusterInput) (*CreateECSClusterOutput, error) {
	r := client.NewRequest(http.MethodPost, "/ocean/aws/ecs/cluster")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := ecsClustersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateECSClusterOutput)
	if len(gs) > 0 {
		output.Cluster = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) ReadECSCluster(ctx context.Context, input *ReadECSClusterInput) (*ReadECSClusterOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/ecs/cluster/{clusterId}", uritemplates.Values{
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

	gs, err := ecsClustersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadECSClusterOutput)
	if len(gs) > 0 {
		output.Cluster = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateECSCluster(ctx context.Context, input *UpdateECSClusterInput) (*UpdateECSClusterOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/ecs/cluster/{clusterId}", uritemplates.Values{
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

	gs, err := ecsClustersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(UpdateECSClusterOutput)
	if len(gs) > 0 {
		output.Cluster = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) DeleteECSCluster(ctx context.Context, input *DeleteECSClusterInput) (*DeleteECSClusterOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/ecs/cluster/{clusterId}", uritemplates.Values{
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

	return &DeleteECSClusterOutput{}, nil
}

func (s *ServiceOp) RollECS(ctx context.Context, input *ECSRollClusterInput) (*ECSRollClusterOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/ecs/cluster/{clusterId}/roll", uritemplates.Values{
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

	rs, err := ecsRollStatusesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ECSRollClusterOutput)
	if len(rs) > 0 {
		output.RollClusterStatus = rs[0]
	}

	return output, nil
}

// region Cluster

func (o ECSCluster) MarshalJSON() ([]byte, error) {
	type noMethod ECSCluster
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSCluster) SetId(v *string) *ECSCluster {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *ECSCluster) SetName(v *string) *ECSCluster {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *ECSCluster) SetClusterName(v *string) *ECSCluster {
	if o.ClusterName = v; o.ClusterName == nil {
		o.nullFields = append(o.nullFields, "ClusterName")
	}
	return o
}

func (o *ECSCluster) SetRegion(v *string) *ECSCluster {
	if o.Region = v; o.Region == nil {
		o.nullFields = append(o.nullFields, "Region")
	}
	return o
}

func (o *ECSCluster) SetECSStrategy(v *ECSStrategy) *ECSCluster {
	if o.Strategy = v; o.Strategy == nil {
		o.nullFields = append(o.nullFields, "Strategy")
	}
	return o
}

func (o ECSCapacity) MarshalJSON() ([]byte, error) {
	type noMethod ECSCapacity
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSCapacity) SetMinimum(v *int) *ECSCapacity {
	if o.Minimum = v; o.Minimum == nil {
		o.nullFields = append(o.nullFields, "Minimum")
	}
	return o
}

func (o *ECSCapacity) SetMaximum(v *int) *ECSCapacity {
	if o.Maximum = v; o.Maximum == nil {
		o.nullFields = append(o.nullFields, "Maximum")
	}
	return o
}

func (o *ECSCapacity) SetTarget(v *int) *ECSCapacity {
	if o.Target = v; o.Target == nil {
		o.nullFields = append(o.nullFields, "Target")
	}
	return o
}

func (o *ECSCluster) SetCapacity(v *ECSCapacity) *ECSCluster {
	if o.Capacity = v; o.Capacity == nil {
		o.nullFields = append(o.nullFields, "Capacity")
	}
	return o
}

func (o *ECSCluster) SetCompute(v *ECSCompute) *ECSCluster {
	if o.Compute = v; o.Compute == nil {
		o.nullFields = append(o.nullFields, "Compute")
	}
	return o
}

func (o *ECSCluster) SetAutoScaler(v *ECSAutoScaler) *ECSCluster {
	if o.AutoScaler = v; o.AutoScaler == nil {
		o.nullFields = append(o.nullFields, "AutoScaler")
	}
	return o
}

func (o *ECSCluster) SetScheduling(v *ECSScheduling) *ECSCluster {
	if o.Scheduling = v; o.Scheduling == nil {
		o.nullFields = append(o.nullFields, "Scheduling")
	}
	return o
}

// endregion

// region Scheduling

func (o ECSScheduling) MarshalJSON() ([]byte, error) {
	type noMethod ECSScheduling
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSScheduling) SetTasks(v []*ECSTask) *ECSScheduling {
	if o.Tasks = v; o.Tasks == nil {
		o.nullFields = append(o.nullFields, "Tasks")
	}
	return o
}

func (o *ECSScheduling) SetShutdownHours(v *ECSShutdownHours) *ECSScheduling {
	if o.ShutdownHours = v; o.ShutdownHours == nil {
		o.nullFields = append(o.nullFields, "ShutdownHours")
	}
	return o
}

// endregion

// region ShutdownHours

func (o ECSShutdownHours) MarshalJSON() ([]byte, error) {
	type noMethod ECSShutdownHours
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSShutdownHours) SetIsEnabled(v *bool) *ECSShutdownHours {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}

func (o *ECSShutdownHours) SetTimeWindows(v []string) *ECSShutdownHours {
	if o.TimeWindows = v; o.TimeWindows == nil {
		o.nullFields = append(o.nullFields, "TimeWindows")
	}
	return o
}

// endregion

// region Tasks

func (o ECSTask) MarshalJSON() ([]byte, error) {
	type noMethod ECSTask
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSTask) SetIsEnabled(v *bool) *ECSTask {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}

func (o *ECSTask) SetType(v *string) *ECSTask {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o *ECSTask) SetCronExpression(v *string) *ECSTask {
	if o.CronExpression = v; o.CronExpression == nil {
		o.nullFields = append(o.nullFields, "CronExpression")
	}
	return o
}

// endregion

// region Compute

func (o ECSCompute) MarshalJSON() ([]byte, error) {
	type noMethod ECSCompute
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSCompute) SetInstanceTypes(v *ECSInstanceTypes) *ECSCompute {
	if o.InstanceTypes = v; o.InstanceTypes == nil {
		o.nullFields = append(o.nullFields, "InstanceTypes")
	}
	return o
}

func (o *ECSCompute) SetLaunchSpecification(v *ECSLaunchSpecification) *ECSCompute {
	if o.LaunchSpecification = v; o.LaunchSpecification == nil {
		o.nullFields = append(o.nullFields, "LaunchSpecification")
	}
	return o
}

func (o *ECSCompute) SetSubnetIDs(v []string) *ECSCompute {
	if o.SubnetIDs = v; o.SubnetIDs == nil {
		o.nullFields = append(o.nullFields, "SubnetIDs")
	}
	return o
}

// endregion

// region Strategy

func (o ECSStrategy) MarshalJSON() ([]byte, error) {
	type noMethod ECSStrategy
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSStrategy) SetDrainingTimeout(v *int) *ECSStrategy {
	if o.DrainingTimeout = v; o.DrainingTimeout == nil {
		o.nullFields = append(o.nullFields, "DrainingTimeout")
	}
	return o
}

func (o *ECSStrategy) SetUtilizeReservedInstances(v *bool) *ECSStrategy {
	if o.UtilizeReservedInstances = v; o.UtilizeReservedInstances == nil {
		o.nullFields = append(o.nullFields, "UtilizeReservedInstances")
	}
	return o
}

// endregion

// region InstanceTypes

func (o ECSInstanceTypes) MarshalJSON() ([]byte, error) {
	type noMethod ECSInstanceTypes
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSInstanceTypes) SetWhitelist(v []string) *ECSInstanceTypes {
	if o.Whitelist = v; o.Whitelist == nil {
		o.nullFields = append(o.nullFields, "Whitelist")
	}
	return o
}

// endregion

// region LaunchSpecification

func (o ECSLaunchSpecification) MarshalJSON() ([]byte, error) {
	type noMethod ECSLaunchSpecification
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSLaunchSpecification) SetAssociatePublicIPAddress(v *bool) *ECSLaunchSpecification {
	if o.AssociatePublicIPAddress = v; o.AssociatePublicIPAddress == nil {
		o.nullFields = append(o.nullFields, "AssociatePublicIPAddress")
	}
	return o
}

func (o *ECSLaunchSpecification) SetSecurityGroupIDs(v []string) *ECSLaunchSpecification {
	if o.SecurityGroupIDs = v; o.SecurityGroupIDs == nil {
		o.nullFields = append(o.nullFields, "SecurityGroupIDs")
	}
	return o
}

func (o *ECSLaunchSpecification) SetImageId(v *string) *ECSLaunchSpecification {
	if o.ImageID = v; o.ImageID == nil {
		o.nullFields = append(o.nullFields, "ImageID")
	}
	return o
}

func (o *ECSLaunchSpecification) SetKeyPair(v *string) *ECSLaunchSpecification {
	if o.KeyPair = v; o.KeyPair == nil {
		o.nullFields = append(o.nullFields, "KeyPair")
	}
	return o
}

func (o *ECSLaunchSpecification) SetUserData(v *string) *ECSLaunchSpecification {
	if o.UserData = v; o.UserData == nil {
		o.nullFields = append(o.nullFields, "UserData")
	}
	return o
}

func (o *ECSLaunchSpecification) SetIAMInstanceProfile(v *ECSIAMInstanceProfile) *ECSLaunchSpecification {
	if o.IAMInstanceProfile = v; o.IAMInstanceProfile == nil {
		o.nullFields = append(o.nullFields, "IAMInstanceProfile")
	}
	return o
}

func (o *ECSLaunchSpecification) SetTags(v []*Tag) *ECSLaunchSpecification {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

func (o *ECSLaunchSpecification) SetMonitoring(v *bool) *ECSLaunchSpecification {
	if o.Monitoring = v; o.Monitoring == nil {
		o.nullFields = append(o.nullFields, "Monitoring")
	}
	return o
}

func (o *ECSLaunchSpecification) SetEBSOptimized(v *bool) *ECSLaunchSpecification {
	if o.EBSOptimized = v; o.EBSOptimized == nil {
		o.nullFields = append(o.nullFields, "EBSOptimized")
	}
	return o
}

// endregion

// region IAMInstanceProfile

func (o ECSIAMInstanceProfile) MarshalJSON() ([]byte, error) {
	type noMethod ECSIAMInstanceProfile
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSIAMInstanceProfile) SetArn(v *string) *ECSIAMInstanceProfile {
	if o.ARN = v; o.ARN == nil {
		o.nullFields = append(o.nullFields, "ARN")
	}
	return o
}

func (o *ECSIAMInstanceProfile) SetName(v *string) *ECSIAMInstanceProfile {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

// endregion

// region AutoScaler

func (o ECSAutoScaler) MarshalJSON() ([]byte, error) {
	type noMethod ECSAutoScaler
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSAutoScaler) SetIsEnabled(v *bool) *ECSAutoScaler {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}

func (o *ECSAutoScaler) SetIsAutoConfig(v *bool) *ECSAutoScaler {
	if o.IsAutoConfig = v; o.IsAutoConfig == nil {
		o.nullFields = append(o.nullFields, "IsAutoConfig")
	}
	return o
}

func (o *ECSAutoScaler) SetCooldown(v *int) *ECSAutoScaler {
	if o.Cooldown = v; o.Cooldown == nil {
		o.nullFields = append(o.nullFields, "Cooldown")
	}
	return o
}

func (o *ECSAutoScaler) SetHeadroom(v *ECSAutoScalerHeadroom) *ECSAutoScaler {
	if o.Headroom = v; o.Headroom == nil {
		o.nullFields = append(o.nullFields, "Headroom")
	}
	return o
}

func (o *ECSAutoScaler) SetResourceLimits(v *ECSAutoScalerResourceLimits) *ECSAutoScaler {
	if o.ResourceLimits = v; o.ResourceLimits == nil {
		o.nullFields = append(o.nullFields, "ResourceLimits")
	}
	return o
}

func (o *ECSAutoScaler) SetDown(v *ECSAutoScalerDown) *ECSAutoScaler {
	if o.Down = v; o.Down == nil {
		o.nullFields = append(o.nullFields, "Down")
	}
	return o
}

// endregion

// region AutoScalerHeadroom

func (o ECSAutoScalerHeadroom) MarshalJSON() ([]byte, error) {
	type noMethod ECSAutoScalerHeadroom
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSAutoScalerHeadroom) SetCPUPerUnit(v *int) *ECSAutoScalerHeadroom {
	if o.CPUPerUnit = v; o.CPUPerUnit == nil {
		o.nullFields = append(o.nullFields, "CPUPerUnit")
	}
	return o
}

func (o *ECSAutoScalerHeadroom) SetMemoryPerUnit(v *int) *ECSAutoScalerHeadroom {
	if o.MemoryPerUnit = v; o.MemoryPerUnit == nil {
		o.nullFields = append(o.nullFields, "MemoryPerUnit")
	}
	return o
}

func (o *ECSAutoScalerHeadroom) SetNumOfUnits(v *int) *ECSAutoScalerHeadroom {
	if o.NumOfUnits = v; o.NumOfUnits == nil {
		o.nullFields = append(o.nullFields, "NumOfUnits")
	}
	return o
}

// endregion

// region AutoScalerResourceLimits

func (o ECSAutoScalerResourceLimits) MarshalJSON() ([]byte, error) {
	type noMethod ECSAutoScalerResourceLimits
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSAutoScalerResourceLimits) SetMaxVCPU(v *int) *ECSAutoScalerResourceLimits {
	if o.MaxVCPU = v; o.MaxVCPU == nil {
		o.nullFields = append(o.nullFields, "MaxVCPU")
	}
	return o
}

func (o *ECSAutoScalerResourceLimits) SetMaxMemoryGiB(v *int) *ECSAutoScalerResourceLimits {
	if o.MaxMemoryGiB = v; o.MaxMemoryGiB == nil {
		o.nullFields = append(o.nullFields, "MaxMemoryGiB")
	}
	return o
}

// endregion

// region AutoScalerDown

func (o ECSAutoScalerDown) MarshalJSON() ([]byte, error) {
	type noMethod ECSAutoScalerDown
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSAutoScalerDown) SetMaxScaleDownPercentage(v *int) *ECSAutoScalerDown {
	if o.MaxScaleDownPercentage = v; o.MaxScaleDownPercentage == nil {
		o.nullFields = append(o.nullFields, "MaxScaleDownPercentage")
	}
	return o
}

// endregion
