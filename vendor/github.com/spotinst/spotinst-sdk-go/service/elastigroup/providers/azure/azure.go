package azure

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

type Group struct {
	ID          *string      `json:"id,omitempty"`
	Name        *string      `json:"name,omitempty"`
	Description *string      `json:"description,omitempty"`
	Capacity    *Capacity    `json:"capacity,omitempty"`
	Compute     *Compute     `json:"compute,omitempty"`
	Strategy    *Strategy    `json:"strategy,omitempty"`
	Scaling     *Scaling     `json:"scaling,omitempty"`
	Scheduling  *Scheduling  `json:"scheduling,omitempty"`
	Integration *Integration `json:"thirdPartiesIntegration,omitempty"`

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

type Scheduling struct {
	Tasks []*ScheduledTask `json:"tasks,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Integration struct {
	Rancher *RancherIntegration `json:"rancher,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type RancherIntegration struct {
	MasterHost *string `json:"masterHost,omitempty"`
	AccessKey  *string `json:"accessKey,omitempty"`
	SecretKey  *string `json:"secretKey,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ScheduledTask struct {
	IsEnabled            *bool   `json:"isEnabled,omitempty"`
	Frequency            *string `json:"frequency,omitempty"`
	CronExpression       *string `json:"cronExpression,omitempty"`
	TaskType             *string `json:"taskType,omitempty"`
	ScaleTargetCapacity  *int    `json:"scaleTargetCapacity,omitempty"`
	ScaleMinCapacity     *int    `json:"scaleMinCapacity,omitempty"`
	ScaleMaxCapacity     *int    `json:"scaleMaxCapacity,omitempty"`
	BatchSizePercentage  *int    `json:"batchSizePercentage,omitempty"`
	GracePeriod          *int    `json:"gracePeriod,omitempty"`
	Adjustment           *int    `json:"adjustment,omitempty"`
	AdjustmentPercentage *int    `json:"adjustmentPercentage,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Scaling struct {
	Up   []*ScalingPolicy `json:"up,omitempty"`
	Down []*ScalingPolicy `json:"down,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ScalingPolicy struct {
	PolicyName        *string      `json:"policyName,omitempty"`
	MetricName        *string      `json:"metricName,omitempty"`
	Statistic         *string      `json:"statistic,omitempty"`
	Unit              *string      `json:"unit,omitempty"`
	Threshold         *float64     `json:"threshold,omitempty"`
	Adjustment        *int         `json:"adjustment,omitempty"`
	MinTargetCapacity *int         `json:"minTargetCapacity,omitempty"`
	MaxTargetCapacity *int         `json:"maxTargetCapacity,omitempty"`
	Namespace         *string      `json:"namespace,omitempty"`
	EvaluationPeriods *int         `json:"evaluationPeriods,omitempty"`
	Period            *int         `json:"period,omitempty"`
	Cooldown          *int         `json:"cooldown,omitempty"`
	Operator          *string      `json:"operator,omitempty"`
	Dimensions        []*Dimension `json:"dimensions,omitempty"`
	Action            *Action      `json:"action,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Action struct {
	Type              *string `json:"type,omitempty"`
	Adjustment        *string `json:"adjustment,omitempty"`
	MinTargetCapacity *string `json:"minTargetCapacity,omitempty"`
	MaxTargetCapacity *string `json:"maxTargetCapacity,omitempty"`
	Maximum           *string `json:"maximum,omitempty"`
	Minimum           *string `json:"minimum,omitempty"`
	Target            *string `json:"target,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Dimension struct {
	Name  *string `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Strategy struct {
	LowPriorityPercentage *int      `json:"lowPriorityPercentage,omitempty"`
	DedicatedCount        *int      `json:"dedicatedCount,omitempty"`
	DrainingTimeout       *int      `json:"drainingTimeout,omitempty"`
	Signals               []*Signal `json:"signals,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Signal struct {
	Name    *string `json:"name,omitempty"`
	Timeout *int    `json:"timeout,omitempty"`

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
	Region              *string              `json:"region,omitempty"`
	Product             *string              `json:"product,omitempty"`
	ResourceGroupName   *string              `json:"resourceGroupName,omitempty"`
	VMSize              *VMSize              `json:"vmSizes,omitempty"`
	LaunchSpecification *LaunchSpecification `json:"launchSpecification,omitempty"`
	Health              *Health              `json:"health,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type VMSize struct {
	Dedicated   []string `json:"dedicatedSizes,omitempty"`
	LowPriority []string `json:"lowPrioritySizes,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type LaunchSpecification struct {
	LoadBalancersConfig *LoadBalancersConfig `json:"loadBalancersConfig,omitempty"`
	Image               *Image               `json:"image,omitempty"`
	UserData            *UserData            `json:"userData,omitempty"`
	Storage             *Storage             `json:"storage,omitempty"`
	Network             *Network             `json:"network,omitempty"`
	SSHPublicKey        *string              `json:"sshPublicKey,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type LoadBalancersConfig struct {
	LoadBalancers []*LoadBalancer `json:"loadBalancers,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type LoadBalancer struct {
	BalancerID  *string `json:"balancerId,omitempty"`
	TargetSetID *string `json:"targetSetId,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Image struct {
	Custom    *CustomImage `json:"customImage,omitempty"`
	Publisher *string      `json:"publisher,omitempty"`
	Offer     *string      `json:"offer,omitempty"`
	SKU       *string      `json:"sku,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type CustomImage struct {
	ImageURIs []string `json:"imageUris,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type UserData struct {
	CommandLine   *string         `json:"commandLine,omitempty"`
	ResourceFiles []*ResourceFile `json:"resourceFiles,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ResourceFile struct {
	URL        *string `json:"resourceFileUrl,omitempty"`
	TargetPath *string `json:"resourceFileTargetPath,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Storage struct {
	AccountName *string `json:"storageAccountName,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Network struct {
	VirtualNetworkName *string `json:"virtualNetworkName,omitempty"`
	SubnetID           *string `json:"subnetId,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Health struct {
	HealthCheckType *string `json:"healthCheckType,omitempty"`
	AutoHealing     *bool   `json:"autoHealing,omitempty"`
	GracePeriod     *int    `json:"gracePeriod,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Node struct {
	ID        *string    `json:"id,omitempty"`
	VMSize    *string    `json:"vmSize,omitempty"`
	State     *string    `json:"state,omitempty"`
	LifeCycle *string    `json:"lifeCycle,omitempty"`
	Region    *string    `json:"region,omitempty"`
	IPAddress *string    `json:"ipAddress,omitempty"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
}

type RollStrategy struct {
	Action               *string `json:"action,omitempty"`
	ShouldDrainInstances *bool   `json:"shouldDrainInstances,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ListGroupsInput struct{}

type ListGroupsOutput struct {
	Groups []*Group `json:"groups,omitempty"`
}

type CreateGroupInput struct {
	Group *Group `json:"group,omitempty"`
}

type CreateGroupOutput struct {
	Group *Group `json:"group,omitempty"`
}

type ReadGroupInput struct {
	GroupID *string `json:"groupId,omitempty"`
}

type ReadGroupOutput struct {
	Group *Group `json:"group,omitempty"`
}

type UpdateGroupInput struct {
	Group *Group `json:"group,omitempty"`
}

type UpdateGroupOutput struct {
	Group *Group `json:"group,omitempty"`
}

type DeleteGroupInput struct {
	GroupID *string `json:"groupId,omitempty"`
}

type DeleteGroupOutput struct{}

type StatusGroupInput struct {
	GroupID *string `json:"groupId,omitempty"`
}

type StatusGroupOutput struct {
	Nodes []*Node `json:"nodes,omitempty"`
}

type DetachGroupInput struct {
	GroupID                       *string  `json:"groupId,omitempty"`
	InstanceIDs                   []string `json:"instancesToDetach,omitempty"`
	ShouldDecrementTargetCapacity *bool    `json:"shouldDecrementTargetCapacity,omitempty"`
	ShouldTerminateInstances      *bool    `json:"shouldTerminateInstances,omitempty"`
	DrainingTimeout               *int     `json:"drainingTimeout,omitempty"`
}

type DetachGroupOutput struct{}

type RollGroupInput struct {
	GroupID             *string       `json:"groupId,omitempty"`
	BatchSizePercentage *int          `json:"batchSizePercentage,omitempty"`
	GracePeriod         *int          `json:"gracePeriod,omitempty"`
	HealthCheckType     *string       `json:"healthCheckType,omitempty"`
	Strategy            *RollStrategy `json:"strategy,omitempty"`
}

type RollGroupOutput struct{}

func groupFromJSON(in []byte) (*Group, error) {
	b := new(Group)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

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

func groupsFromHttpResponse(resp *http.Response) ([]*Group, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return groupsFromJSON(body)
}

func nodeFromJSON(in []byte) (*Node, error) {
	b := new(Node)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func nodesFromJSON(in []byte) ([]*Node, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Node, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := nodeFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func nodesFromHttpResponse(resp *http.Response) ([]*Node, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return nodesFromJSON(body)
}

func (s *ServiceOp) List(ctx context.Context, input *ListGroupsInput) (*ListGroupsOutput, error) {
	r := client.NewRequest(http.MethodGet, "/compute/azure/group")
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

func (s *ServiceOp) Create(ctx context.Context, input *CreateGroupInput) (*CreateGroupOutput, error) {
	r := client.NewRequest(http.MethodPost, "/compute/azure/group")
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

func (s *ServiceOp) Read(ctx context.Context, input *ReadGroupInput) (*ReadGroupOutput, error) {
	path, err := uritemplates.Expand("/compute/azure/group/{groupId}", uritemplates.Values{
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

func (s *ServiceOp) Update(ctx context.Context, input *UpdateGroupInput) (*UpdateGroupOutput, error) {
	path, err := uritemplates.Expand("/compute/azure/group/{groupId}", uritemplates.Values{
		"groupId": spotinst.StringValue(input.Group.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
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

func (s *ServiceOp) Delete(ctx context.Context, input *DeleteGroupInput) (*DeleteGroupOutput, error) {
	path, err := uritemplates.Expand("/compute/azure/group/{groupId}", uritemplates.Values{
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

func (s *ServiceOp) Status(ctx context.Context, input *StatusGroupInput) (*StatusGroupOutput, error) {
	path, err := uritemplates.Expand("/compute/azure/group/{groupId}/status", uritemplates.Values{
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

	ns, err := nodesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &StatusGroupOutput{Nodes: ns}, nil
}

func (s *ServiceOp) Detach(ctx context.Context, input *DetachGroupInput) (*DetachGroupOutput, error) {
	path, err := uritemplates.Expand("/compute/azure/group/{groupId}/detachNodes", uritemplates.Values{
		"groupId": spotinst.StringValue(input.GroupID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.GroupID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DetachGroupOutput{}, nil
}

func (s *ServiceOp) Roll(ctx context.Context, input *RollGroupInput) (*RollGroupOutput, error) {
	path, err := uritemplates.Expand("/compute/azure/group/{groupId}/roll", uritemplates.Values{
		"groupId": spotinst.StringValue(input.GroupID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.GroupID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &RollGroupOutput{}, nil
}

// region Group

func (o *Group) MarshalJSON() ([]byte, error) {
	type noMethod Group
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Group) SetId(v *string) *Group {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *Group) SetName(v *string) *Group {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *Group) SetDescription(v *string) *Group {
	if o.Description = v; o.Description == nil {
		o.nullFields = append(o.nullFields, "Description")
	}
	return o
}

func (o *Group) SetCapacity(v *Capacity) *Group {
	if o.Capacity = v; o.Capacity == nil {
		o.nullFields = append(o.nullFields, "Capacity")
	}
	return o
}

func (o *Group) SetCompute(v *Compute) *Group {
	if o.Compute = v; o.Compute == nil {
		o.nullFields = append(o.nullFields, "Compute")
	}
	return o
}

func (o *Group) SetStrategy(v *Strategy) *Group {
	if o.Strategy = v; o.Strategy == nil {
		o.nullFields = append(o.nullFields, "Strategy")
	}
	return o
}

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

func (o *Group) SetIntegration(v *Integration) *Group {
	if o.Integration = v; o.Integration == nil {
		o.nullFields = append(o.nullFields, "Integration")
	}
	return o
}

// endregion

// region Scheduling

func (o *Scheduling) MarshalJSON() ([]byte, error) {
	type noMethod Scheduling
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Scheduling) SetTasks(v []*ScheduledTask) *Scheduling {
	if o.Tasks = v; o.Tasks == nil {
		o.nullFields = append(o.nullFields, "Tasks")
	}
	return o
}

// endregion

// region Integration

func (o *Integration) MarshalJSON() ([]byte, error) {
	type noMethod Integration
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Integration) SetRancher(v *RancherIntegration) *Integration {
	if o.Rancher = v; o.Rancher == nil {
		o.nullFields = append(o.nullFields, "Rancher")
	}
	return o
}

// endregion

// region RancherIntegration

func (o *RancherIntegration) MarshalJSON() ([]byte, error) {
	type noMethod RancherIntegration
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *RancherIntegration) SetMasterHost(v *string) *RancherIntegration {
	if o.MasterHost = v; o.MasterHost == nil {
		o.nullFields = append(o.nullFields, "MasterHost")
	}
	return o
}

func (o *RancherIntegration) SetAccessKey(v *string) *RancherIntegration {
	if o.AccessKey = v; o.AccessKey == nil {
		o.nullFields = append(o.nullFields, "AccessKey")
	}
	return o
}

func (o *RancherIntegration) SetSecretKey(v *string) *RancherIntegration {
	if o.SecretKey = v; o.SecretKey == nil {
		o.nullFields = append(o.nullFields, "SecretKey")
	}
	return o
}

// endregion

// region ScheduledTask

func (o *ScheduledTask) MarshalJSON() ([]byte, error) {
	type noMethod ScheduledTask
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ScheduledTask) SetIsEnabled(v *bool) *ScheduledTask {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}

func (o *ScheduledTask) SetFrequency(v *string) *ScheduledTask {
	if o.Frequency = v; o.Frequency == nil {
		o.nullFields = append(o.nullFields, "Frequency")
	}
	return o
}

func (o *ScheduledTask) SetCronExpression(v *string) *ScheduledTask {
	if o.CronExpression = v; o.CronExpression == nil {
		o.nullFields = append(o.nullFields, "CronExpression")
	}
	return o
}

func (o *ScheduledTask) SetTaskType(v *string) *ScheduledTask {
	if o.TaskType = v; o.TaskType == nil {
		o.nullFields = append(o.nullFields, "TaskType")
	}
	return o
}

func (o *ScheduledTask) SetScaleTargetCapacity(v *int) *ScheduledTask {
	if o.ScaleTargetCapacity = v; o.ScaleTargetCapacity == nil {
		o.nullFields = append(o.nullFields, "ScaleTargetCapacity")
	}
	return o
}

func (o *ScheduledTask) SetScaleMinCapacity(v *int) *ScheduledTask {
	if o.ScaleMinCapacity = v; o.ScaleMinCapacity == nil {
		o.nullFields = append(o.nullFields, "ScaleMinCapacity")
	}
	return o
}

func (o *ScheduledTask) SetScaleMaxCapacity(v *int) *ScheduledTask {
	if o.ScaleMaxCapacity = v; o.ScaleMaxCapacity == nil {
		o.nullFields = append(o.nullFields, "ScaleMaxCapacity")
	}
	return o
}

func (o *ScheduledTask) SetBatchSizePercentage(v *int) *ScheduledTask {
	if o.BatchSizePercentage = v; o.BatchSizePercentage == nil {
		o.nullFields = append(o.nullFields, "BatchSizePercentage")
	}
	return o
}

func (o *ScheduledTask) SetGracePeriod(v *int) *ScheduledTask {
	if o.GracePeriod = v; o.GracePeriod == nil {
		o.nullFields = append(o.nullFields, "GracePeriod")
	}
	return o
}

func (o *ScheduledTask) SetAdjustment(v *int) *ScheduledTask {
	if o.Adjustment = v; o.Adjustment == nil {
		o.nullFields = append(o.nullFields, "Adjustment")
	}
	return o
}

func (o *ScheduledTask) SetAdjustmentPercentage(v *int) *ScheduledTask {
	if o.AdjustmentPercentage = v; o.AdjustmentPercentage == nil {
		o.nullFields = append(o.nullFields, "AdjustmentPercentage")
	}
	return o
}

// endregion

// region Scaling

func (o *Scaling) MarshalJSON() ([]byte, error) {
	type noMethod Scaling
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Scaling) SetUp(v []*ScalingPolicy) *Scaling {
	if o.Up = v; o.Up == nil {
		o.nullFields = append(o.nullFields, "Up")
	}
	return o
}

func (o *Scaling) SetDown(v []*ScalingPolicy) *Scaling {
	if o.Down = v; o.Down == nil {
		o.nullFields = append(o.nullFields, "Down")
	}
	return o
}

// endregion

// region ScalingPolicy

func (o *ScalingPolicy) MarshalJSON() ([]byte, error) {
	type noMethod ScalingPolicy
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ScalingPolicy) SetPolicyName(v *string) *ScalingPolicy {
	if o.PolicyName = v; o.PolicyName == nil {
		o.nullFields = append(o.nullFields, "PolicyName")
	}
	return o
}

func (o *ScalingPolicy) SetMetricName(v *string) *ScalingPolicy {
	if o.MetricName = v; o.MetricName == nil {
		o.nullFields = append(o.nullFields, "MetricName")
	}
	return o
}

func (o *ScalingPolicy) SetStatistic(v *string) *ScalingPolicy {
	if o.Statistic = v; o.Statistic == nil {
		o.nullFields = append(o.nullFields, "Statistic")
	}
	return o
}

func (o *ScalingPolicy) SetUnit(v *string) *ScalingPolicy {
	if o.Unit = v; o.Unit == nil {
		o.nullFields = append(o.nullFields, "Unit")
	}
	return o
}

func (o *ScalingPolicy) SetThreshold(v *float64) *ScalingPolicy {
	if o.Threshold = v; o.Threshold == nil {
		o.nullFields = append(o.nullFields, "Threshold")
	}
	return o
}

func (o *ScalingPolicy) SetAdjustment(v *int) *ScalingPolicy {
	if o.Adjustment = v; o.Adjustment == nil {
		o.nullFields = append(o.nullFields, "Adjustment")
	}
	return o
}

func (o *ScalingPolicy) SetMinTargetCapacity(v *int) *ScalingPolicy {
	if o.MinTargetCapacity = v; o.MinTargetCapacity == nil {
		o.nullFields = append(o.nullFields, "MinTargetCapacity")
	}
	return o
}

func (o *ScalingPolicy) SetMaxTargetCapacity(v *int) *ScalingPolicy {
	if o.MaxTargetCapacity = v; o.MaxTargetCapacity == nil {
		o.nullFields = append(o.nullFields, "MaxTargetCapacity")
	}
	return o
}

func (o *ScalingPolicy) SetNamespace(v *string) *ScalingPolicy {
	if o.Namespace = v; o.Namespace == nil {
		o.nullFields = append(o.nullFields, "Namespace")
	}
	return o
}

func (o *ScalingPolicy) SetEvaluationPeriods(v *int) *ScalingPolicy {
	if o.EvaluationPeriods = v; o.EvaluationPeriods == nil {
		o.nullFields = append(o.nullFields, "EvaluationPeriods")
	}
	return o
}

func (o *ScalingPolicy) SetPeriod(v *int) *ScalingPolicy {
	if o.Period = v; o.Period == nil {
		o.nullFields = append(o.nullFields, "Period")
	}
	return o
}

func (o *ScalingPolicy) SetCooldown(v *int) *ScalingPolicy {
	if o.Cooldown = v; o.Cooldown == nil {
		o.nullFields = append(o.nullFields, "Cooldown")
	}
	return o
}

func (o *ScalingPolicy) SetOperator(v *string) *ScalingPolicy {
	if o.Operator = v; o.Operator == nil {
		o.nullFields = append(o.nullFields, "Operator")
	}
	return o
}

func (o *ScalingPolicy) SetDimensions(v []*Dimension) *ScalingPolicy {
	if o.Dimensions = v; o.Dimensions == nil {
		o.nullFields = append(o.nullFields, "Dimensions")
	}
	return o
}

func (o *ScalingPolicy) SetAction(v *Action) *ScalingPolicy {
	if o.Action = v; o.Action == nil {
		o.nullFields = append(o.nullFields, "Action")
	}
	return o
}

// endregion

// region Action

func (o *Action) MarshalJSON() ([]byte, error) {
	type noMethod Action
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Action) SetType(v *string) *Action {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o *Action) SetAdjustment(v *string) *Action {
	if o.Adjustment = v; o.Adjustment == nil {
		o.nullFields = append(o.nullFields, "Adjustment")
	}
	return o
}

func (o *Action) SetMinTargetCapacity(v *string) *Action {
	if o.MinTargetCapacity = v; o.MinTargetCapacity == nil {
		o.nullFields = append(o.nullFields, "MinTargetCapacity")
	}
	return o
}

func (o *Action) SetMaxTargetCapacity(v *string) *Action {
	if o.MaxTargetCapacity = v; o.MaxTargetCapacity == nil {
		o.nullFields = append(o.nullFields, "MaxTargetCapacity")
	}
	return o
}

func (o *Action) SetMaximum(v *string) *Action {
	if o.Maximum = v; o.Maximum == nil {
		o.nullFields = append(o.nullFields, "Maximum")
	}
	return o
}

func (o *Action) SetMinimum(v *string) *Action {
	if o.Minimum = v; o.Minimum == nil {
		o.nullFields = append(o.nullFields, "Minimum")
	}
	return o
}

func (o *Action) SetTarget(v *string) *Action {
	if o.Target = v; o.Target == nil {
		o.nullFields = append(o.nullFields, "Target")
	}
	return o
}

// endregion

// region Dimension

func (o *Dimension) MarshalJSON() ([]byte, error) {
	type noMethod Dimension
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Dimension) SetName(v *string) *Dimension {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *Dimension) SetValue(v *string) *Dimension {
	if o.Value = v; o.Value == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}

// endregion

// region Strategy

func (o *Strategy) MarshalJSON() ([]byte, error) {
	type noMethod Strategy
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Strategy) SetLowPriorityPercentage(v *int) *Strategy {
	if o.LowPriorityPercentage = v; o.LowPriorityPercentage == nil {
		o.nullFields = append(o.nullFields, "LowPriorityPercentage")
	}
	return o
}

func (o *Strategy) SetDedicatedCount(v *int) *Strategy {
	if o.DedicatedCount = v; o.DedicatedCount == nil {
		o.nullFields = append(o.nullFields, "DedicatedCount")
	}
	return o
}

func (o *Strategy) SetDrainingTimeout(v *int) *Strategy {
	if o.DrainingTimeout = v; o.DrainingTimeout == nil {
		o.nullFields = append(o.nullFields, "DrainingTimeout")
	}
	return o
}

func (o *Strategy) SetSignals(v []*Signal) *Strategy {
	if o.Signals = v; o.Signals == nil {
		o.nullFields = append(o.nullFields, "Signals")
	}
	return o
}

// endregion

// region Signal

func (o *Signal) MarshalJSON() ([]byte, error) {
	type noMethod Signal
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Signal) SetName(v *string) *Signal {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *Signal) SetTimeout(v *int) *Signal {
	if o.Timeout = v; o.Timeout == nil {
		o.nullFields = append(o.nullFields, "Timeout")
	}
	return o
}

// endregion

// region Capacity

func (o *Capacity) MarshalJSON() ([]byte, error) {
	type noMethod Capacity
	raw := noMethod(*o)
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

func (o *Compute) MarshalJSON() ([]byte, error) {
	type noMethod Compute
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Compute) SetRegion(v *string) *Compute {
	if o.Region = v; o.Region == nil {
		o.nullFields = append(o.nullFields, "Region")
	}
	return o
}

func (o *Compute) SetProduct(v *string) *Compute {
	if o.Product = v; o.Product == nil {
		o.nullFields = append(o.nullFields, "Product")
	}
	return o
}

func (o *Compute) SetResourceGroupName(v *string) *Compute {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}

func (o *Compute) SetVMSize(v *VMSize) *Compute {
	if o.VMSize = v; o.VMSize == nil {
		o.nullFields = append(o.nullFields, "VMSize")
	}
	return o
}

func (o *Compute) SetLaunchSpecification(v *LaunchSpecification) *Compute {
	if o.LaunchSpecification = v; o.LaunchSpecification == nil {
		o.nullFields = append(o.nullFields, "LaunchSpecification")
	}
	return o
}

func (o *Compute) SetHealth(v *Health) *Compute {
	if o.Health = v; o.Health == nil {
		o.nullFields = append(o.nullFields, "Health")
	}
	return o
}

// endregion

// region VMSize

func (o *VMSize) MarshalJSON() ([]byte, error) {
	type noMethod VMSize
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *VMSize) SetDedicated(v []string) *VMSize {
	if o.Dedicated = v; o.Dedicated == nil {
		o.nullFields = append(o.nullFields, "Dedicated")
	}
	return o
}

func (o *VMSize) SetLowPriority(v []string) *VMSize {
	if o.LowPriority = v; o.LowPriority == nil {
		o.nullFields = append(o.nullFields, "LowPriority")
	}
	return o
}

// endregion

// region LaunchSpecification

func (o *LaunchSpecification) MarshalJSON() ([]byte, error) {
	type noMethod LaunchSpecification
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *LaunchSpecification) SetLoadBalancersConfig(v *LoadBalancersConfig) *LaunchSpecification {
	if o.LoadBalancersConfig = v; o.LoadBalancersConfig == nil {
		o.nullFields = append(o.nullFields, "LoadBalancersConfig")
	}
	return o
}

func (o *LaunchSpecification) SetImage(v *Image) *LaunchSpecification {
	if o.Image = v; o.Image == nil {
		o.nullFields = append(o.nullFields, "Image")
	}
	return o
}

func (o *LaunchSpecification) SetUserData(v *UserData) *LaunchSpecification {
	if o.UserData = v; o.UserData == nil {
		o.nullFields = append(o.nullFields, "UserData")
	}
	return o
}

func (o *LaunchSpecification) SetStorage(v *Storage) *LaunchSpecification {
	if o.Storage = v; o.Storage == nil {
		o.nullFields = append(o.nullFields, "Storage")
	}
	return o
}

func (o *LaunchSpecification) SetNetwork(v *Network) *LaunchSpecification {
	if o.Network = v; o.Network == nil {
		o.nullFields = append(o.nullFields, "Network")
	}
	return o
}

func (o *LaunchSpecification) SetSSHPublicKey(v *string) *LaunchSpecification {
	if o.SSHPublicKey = v; o.SSHPublicKey == nil {
		o.nullFields = append(o.nullFields, "SSHPublicKey")
	}
	return o
}

// endregion

// region LoadBalancersConfig

func (o *LoadBalancersConfig) MarshalJSON() ([]byte, error) {
	type noMethod LoadBalancersConfig
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *LoadBalancersConfig) SetLoadBalancers(v []*LoadBalancer) *LoadBalancersConfig {
	if o.LoadBalancers = v; o.LoadBalancers == nil {
		o.nullFields = append(o.nullFields, "LoadBalancers")
	}
	return o
}

// endregion

// region LoadBalancer

func (o *LoadBalancer) MarshalJSON() ([]byte, error) {
	type noMethod LoadBalancer
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *LoadBalancer) SetBalancerId(v *string) *LoadBalancer {
	if o.BalancerID = v; o.BalancerID == nil {
		o.nullFields = append(o.nullFields, "BalancerID")
	}
	return o
}

func (o *LoadBalancer) SetTargetSetId(v *string) *LoadBalancer {
	if o.TargetSetID = v; o.TargetSetID == nil {
		o.nullFields = append(o.nullFields, "TargetSetID")
	}
	return o
}

// endregion

// region Image

func (o *Image) MarshalJSON() ([]byte, error) {
	type noMethod Image
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Image) SetCustom(v *CustomImage) *Image {
	if o.Custom = v; o.Custom == nil {
		o.nullFields = append(o.nullFields, "Custom")
	}
	return o
}

func (o *Image) SetPublisher(v *string) *Image {
	if o.Publisher = v; o.Publisher == nil {
		o.nullFields = append(o.nullFields, "Publisher")
	}
	return o
}

func (o *Image) SetOffer(v *string) *Image {
	if o.Offer = v; o.Offer == nil {
		o.nullFields = append(o.nullFields, "Offer")
	}
	return o
}

func (o *Image) SetSKU(v *string) *Image {
	if o.SKU = v; o.SKU == nil {
		o.nullFields = append(o.nullFields, "SKU")
	}
	return o
}

// endregion

// region CustomImage

func (o *CustomImage) MarshalJSON() ([]byte, error) {
	type noMethod CustomImage
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *CustomImage) SetImageURIs(v []string) *CustomImage {
	if o.ImageURIs = v; o.ImageURIs == nil {
		o.nullFields = append(o.nullFields, "ImageURIs")
	}
	return o
}

// endregion

// region UserData

func (o *UserData) MarshalJSON() ([]byte, error) {
	type noMethod UserData
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *UserData) SetCommandLine(v *string) *UserData {
	if o.CommandLine = v; o.CommandLine == nil {
		o.nullFields = append(o.nullFields, "CommandLine")
	}
	return o
}

func (o *UserData) SetResourceFiles(v []*ResourceFile) *UserData {
	if o.ResourceFiles = v; o.ResourceFiles == nil {
		o.nullFields = append(o.nullFields, "ResourceFiles")
	}
	return o
}

// endregion

// region ResourceFile

func (o *ResourceFile) MarshalJSON() ([]byte, error) {
	type noMethod ResourceFile
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ResourceFile) SetURL(v *string) *ResourceFile {
	if o.URL = v; o.URL == nil {
		o.nullFields = append(o.nullFields, "URL")
	}
	return o
}

func (o *ResourceFile) SetTargetPath(v *string) *ResourceFile {
	if o.TargetPath = v; o.TargetPath == nil {
		o.nullFields = append(o.nullFields, "TargetPath")
	}
	return o
}

// endregion

// region Storage

func (o *Storage) MarshalJSON() ([]byte, error) {
	type noMethod Storage
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Storage) SetAccountName(v *string) *Storage {
	if o.AccountName = v; o.AccountName == nil {
		o.nullFields = append(o.nullFields, "AccountName")
	}
	return o
}

// endregion

// region Network

func (o *Network) MarshalJSON() ([]byte, error) {
	type noMethod Network
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Network) SetVirtualNetworkName(v *string) *Network {
	if o.VirtualNetworkName = v; o.VirtualNetworkName == nil {
		o.nullFields = append(o.nullFields, "VirtualNetworkName")
	}
	return o
}

func (o *Network) SetSubnetId(v *string) *Network {
	if o.SubnetID = v; o.SubnetID == nil {
		o.nullFields = append(o.nullFields, "SubnetID")
	}
	return o
}

// endregion

// region Health

func (o *Health) MarshalJSON() ([]byte, error) {
	type noMethod Health
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Health) SetHealthCheckType(v *string) *Health {
	if o.HealthCheckType = v; o.HealthCheckType == nil {
		o.nullFields = append(o.nullFields, "HealthCheckType")
	}
	return o
}

func (o *Health) SetAutoHealing(v *bool) *Health {
	if o.AutoHealing = v; o.AutoHealing == nil {
		o.nullFields = append(o.nullFields, "AutoHealing")
	}
	return o
}

func (o *Health) SetGracePeriod(v *int) *Health {
	if o.GracePeriod = v; o.GracePeriod == nil {
		o.nullFields = append(o.nullFields, "GracePeriod")
	}
	return o
}

// endregion
