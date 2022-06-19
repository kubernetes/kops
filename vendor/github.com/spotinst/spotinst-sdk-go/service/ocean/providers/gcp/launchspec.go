package gcp

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/jsonutil"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/uritemplates"
)

type LaunchSpec struct {
	ID                     *string                  `json:"id,omitempty"`
	Name                   *string                  `json:"name,omitempty"`
	OceanID                *string                  `json:"oceanId,omitempty"`
	SourceImage            *string                  `json:"sourceImage,omitempty"`
	Metadata               []*Metadata              `json:"metadata,omitempty"`
	Labels                 []*Label                 `json:"labels,omitempty"`
	Taints                 []*Taint                 `json:"taints,omitempty"`
	AutoScale              *AutoScale               `json:"autoScale,omitempty"`
	RestrictScaleDown      *bool                    `json:"restrictScaleDown,omitempty"`
	Strategy               *LaunchSpecStrategy      `json:"strategy,omitempty"`
	RootVolumeSizeInGB     *int                     `json:"rootVolumeSizeInGb,omitempty"`
	RootVolumeType         *string                  `json:"rootVolumeType,omitempty"`
	ShieldedInstanceConfig *ShieldedInstanceConfig  `json:"shieldedInstanceConfig,omitempty"`
	ServiceAccount         *string                  `json:"serviceAccount,omitempty"`
	InstanceTypes          []string                 `json:"instanceTypes,omitempty"`
	Storage                *Storage                 `json:"storage,omitempty"`
	ResourceLimits         *ResourceLimits          `json:"resourceLimits,omitempty"`
	LaunchSpecScheduling   *GKELaunchSpecScheduling `json:"scheduling,omitempty"`
	LaunchSpecTags         []string                 `json:"tags,omitempty"`

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

type Label struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Taint struct {
	Key    *string `json:"key,omitempty"`
	Value  *string `json:"value,omitempty"`
	Effect *string `json:"effect,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AutoScale struct {
	Headrooms              []*AutoScaleHeadroom `json:"headrooms,omitempty"`
	AutoHeadroomPercentage *int                 `json:"autoHeadroomPercentage,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AutoScaleHeadroom struct {
	CPUPerUnit    *int `json:"cpuPerUnit,omitempty"`
	GPUPerUnit    *int `json:"gpuPerUnit,omitempty"`
	MemoryPerUnit *int `json:"memoryPerUnit,omitempty"`
	NumOfUnits    *int `json:"numOfUnits,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type LaunchSpecStrategy struct {
	PreemptiblePercentage *int `json:"preemptiblePercentage,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ShieldedInstanceConfig struct {
	EnableSecureBoot          *bool `json:"enableSecureBoot,omitempty"`
	EnableIntegrityMonitoring *bool `json:"enableIntegrityMonitoring,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Storage struct {
	LocalSSDCount *int `json:"localSsdCount,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ResourceLimits struct {
	MaxInstanceCount *int `json:"maxInstanceCount,omitempty"`
	MinInstanceCount *int `json:"minInstanceCount,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type GKELaunchSpecScheduling struct {
	Tasks []*GKELaunchSpecTask `json:"tasks,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type GKELaunchSpecTask struct {
	IsEnabled      *bool          `json:"isEnabled,omitempty"`
	CronExpression *string        `json:"cronExpression,omitempty"`
	TaskType       *string        `json:"taskType,omitempty"`
	Config         *GKETaskConfig `json:"config,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type GKETaskConfig struct {
	TaskHeadrooms []*GKELaunchSpecTaskHeadroom `json:"headrooms,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type GKELaunchSpecTaskHeadroom struct {
	CPUPerUnit    *int `json:"cpuPerUnit,omitempty"`
	GPUPerUnit    *int `json:"gpuPerUnit,omitempty"`
	MemoryPerUnit *int `json:"memoryPerUnit,omitempty"`
	NumOfUnits    *int `json:"numOfUnits,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ListLaunchSpecsInput struct {
	OceanID *string `json:"oceanId,omitempty"`
}

type ListLaunchSpecsOutput struct {
	LaunchSpecs []*LaunchSpec `json:"launchSpecs,omitempty"`
}

type CreateLaunchSpecInput struct {
	LaunchSpec *LaunchSpec `json:"launchSpec,omitempty"`
}

type CreateLaunchSpecOutput struct {
	LaunchSpec *LaunchSpec `json:"launchSpec,omitempty"`
}

type ReadLaunchSpecInput struct {
	LaunchSpecID *string `json:"launchSpecId,omitempty"`
}

type ReadLaunchSpecOutput struct {
	LaunchSpec *LaunchSpec `json:"launchSpec,omitempty"`
}

type UpdateLaunchSpecInput struct {
	LaunchSpec *LaunchSpec `json:"launchSpec,omitempty"`
}

type UpdateLaunchSpecOutput struct {
	LaunchSpec *LaunchSpec `json:"launchSpec,omitempty"`
}

type DeleteLaunchSpecInput struct {
	LaunchSpecID *string `json:"launchSpecId,omitempty"`
}

type DeleteLaunchSpecOutput struct{}

func launchSpecFromJSON(in []byte) (*LaunchSpec, error) {
	b := new(LaunchSpec)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func launchSpecsFromJSON(in []byte) ([]*LaunchSpec, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*LaunchSpec, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := launchSpecFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func launchSpecsFromHttpResponse(resp *http.Response) ([]*LaunchSpec, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return launchSpecsFromJSON(body)
}

func (s *ServiceOp) ListLaunchSpecs(ctx context.Context, input *ListLaunchSpecsInput) (*ListLaunchSpecsOutput, error) {
	r := client.NewRequest(http.MethodGet, "/ocean/gcp/k8s/launchSpec")

	if input.OceanID != nil {
		r.Params.Set("oceanId", spotinst.StringValue(input.OceanID))
	}

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := launchSpecsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListLaunchSpecsOutput{LaunchSpecs: gs}, nil
}

func (s *ServiceOp) CreateLaunchSpec(ctx context.Context, input *CreateLaunchSpecInput) (*CreateLaunchSpecOutput, error) {
	r := client.NewRequest(http.MethodPost, "/ocean/gcp/k8s/launchSpec")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := launchSpecsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateLaunchSpecOutput)
	if len(gs) > 0 {
		output.LaunchSpec = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) ReadLaunchSpec(ctx context.Context, input *ReadLaunchSpecInput) (*ReadLaunchSpecOutput, error) {
	path, err := uritemplates.Expand("/ocean/gcp/k8s/launchSpec/{launchSpecId}", uritemplates.Values{
		"launchSpecId": spotinst.StringValue(input.LaunchSpecID),
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

	gs, err := launchSpecsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadLaunchSpecOutput)
	if len(gs) > 0 {
		output.LaunchSpec = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateLaunchSpec(ctx context.Context, input *UpdateLaunchSpecInput) (*UpdateLaunchSpecOutput, error) {
	path, err := uritemplates.Expand("/ocean/gcp/k8s/launchSpec/{launchSpecId}", uritemplates.Values{
		"launchSpecId": spotinst.StringValue(input.LaunchSpec.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.LaunchSpec.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := launchSpecsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(UpdateLaunchSpecOutput)
	if len(gs) > 0 {
		output.LaunchSpec = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) DeleteLaunchSpec(ctx context.Context, input *DeleteLaunchSpecInput) (*DeleteLaunchSpecOutput, error) {
	path, err := uritemplates.Expand("/ocean/gcp/k8s/launchSpec/{launchSpecId}", uritemplates.Values{
		"launchSpecId": spotinst.StringValue(input.LaunchSpecID),
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

	return &DeleteLaunchSpecOutput{}, nil
}

func (s *ServiceOp) ImportOceanGKELaunchSpec(ctx context.Context, input *ImportOceanGKELaunchSpecInput) (*ImportOceanGKELaunchSpecOutput, error) {
	r := client.NewRequest(http.MethodPost, "/ocean/gcp/k8s/launchSpec/import")

	r.Params["oceanId"] = []string{spotinst.StringValue(input.OceanId)}
	r.Params["nodePoolName"] = []string{spotinst.StringValue(input.NodePoolName)}

	body := &ImportOceanGKELaunchSpecInput{}
	r.Obj = body

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ls, err := launchSpecsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ImportOceanGKELaunchSpecOutput)
	if len(ls) > 0 {
		output.LaunchSpec = ls[0]
	}

	return output, nil
}

// endregion

// region LaunchSpec

func (o LaunchSpec) MarshalJSON() ([]byte, error) {
	type noMethod LaunchSpec
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *LaunchSpec) SetId(v *string) *LaunchSpec {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *LaunchSpec) SetOceanId(v *string) *LaunchSpec {
	if o.OceanID = v; o.OceanID == nil {
		o.nullFields = append(o.nullFields, "OceanID")
	}
	return o
}

func (o *LaunchSpec) SetName(v *string) *LaunchSpec {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *LaunchSpec) SetSourceImage(v *string) *LaunchSpec {
	if o.SourceImage = v; o.SourceImage == nil {
		o.nullFields = append(o.nullFields, "SourceImage")
	}
	return o
}

func (o *LaunchSpec) SetMetadata(v []*Metadata) *LaunchSpec {
	if o.Metadata = v; o.Metadata == nil {
		o.nullFields = append(o.nullFields, "Metadata")
	}
	return o
}

func (o *LaunchSpec) SetLabels(v []*Label) *LaunchSpec {
	if o.Labels = v; o.Labels == nil {
		o.nullFields = append(o.nullFields, "Labels")
	}
	return o
}

func (o *LaunchSpec) SetTaints(v []*Taint) *LaunchSpec {
	if o.Taints = v; o.Taints == nil {
		o.nullFields = append(o.nullFields, "Taints")
	}
	return o
}

func (o *LaunchSpec) SetAutoScale(v *AutoScale) *LaunchSpec {
	if o.AutoScale = v; o.AutoScale == nil {
		o.nullFields = append(o.nullFields, "AutoScale")
	}
	return o
}

func (o *LaunchSpec) SetRestrictScaleDown(v *bool) *LaunchSpec {
	if o.RestrictScaleDown = v; o.RestrictScaleDown == nil {
		o.nullFields = append(o.nullFields, "RestrictScaleDown")
	}
	return o
}

func (o *LaunchSpec) SetStrategy(v *LaunchSpecStrategy) *LaunchSpec {
	if o.Strategy = v; o.Strategy == nil {
		o.nullFields = append(o.nullFields, "Strategy")
	}
	return o
}

func (o *LaunchSpec) SetRootVolumeSizeInGB(v *int) *LaunchSpec {
	if o.RootVolumeSizeInGB = v; o.RootVolumeSizeInGB == nil {
		o.nullFields = append(o.nullFields, "RootVolumeSizeInGB")
	}
	return o
}

func (o *LaunchSpec) SetRootVolumeType(v *string) *LaunchSpec {
	if o.RootVolumeType = v; o.RootVolumeType == nil {
		o.nullFields = append(o.nullFields, "RootVolumeType")
	}
	return o
}

func (o *LaunchSpec) SetServiceAccount(v *string) *LaunchSpec {
	if o.ServiceAccount = v; o.ServiceAccount == nil {
		o.nullFields = append(o.nullFields, "ServiceAccount")
	}
	return o
}

func (o *LaunchSpec) SetShieldedInstanceConfig(v *ShieldedInstanceConfig) *LaunchSpec {
	if o.ShieldedInstanceConfig = v; o.ShieldedInstanceConfig == nil {
		o.nullFields = append(o.nullFields, "ShieldedInstanceConfig")
	}
	return o
}

func (o *LaunchSpec) SetInstanceTypes(v []string) *LaunchSpec {
	if o.InstanceTypes = v; o.InstanceTypes == nil {
		o.nullFields = append(o.nullFields, "InstanceTypes")
	}
	return o
}

func (o *LaunchSpec) SetStorage(v *Storage) *LaunchSpec {
	if o.Storage = v; o.Storage == nil {
		o.nullFields = append(o.nullFields, "Storage")
	}
	return o
}

func (o *LaunchSpec) SetResourceLimits(v *ResourceLimits) *LaunchSpec {
	if o.ResourceLimits = v; o.ResourceLimits == nil {
		o.nullFields = append(o.nullFields, "ResourceLimits")
	}
	return o
}

func (o *LaunchSpec) SetScheduling(v *GKELaunchSpecScheduling) *LaunchSpec {
	if o.LaunchSpecScheduling = v; o.LaunchSpecScheduling == nil {
		o.nullFields = append(o.nullFields, "GKELaunchSpecScheduling")
	}
	return o
}

func (o *LaunchSpec) SetLaunchSpecTags(v []string) *LaunchSpec {
	if o.LaunchSpecTags = v; o.LaunchSpecTags == nil {
		o.nullFields = append(o.nullFields, "LaunchSpecTags")
	}
	return o
}

// endregion

// region Label

func (o Label) MarshalJSON() ([]byte, error) {
	type noMethod Label
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Label) SetKey(v *string) *Label {
	if o.Key = v; o.Key == nil {
		o.nullFields = append(o.nullFields, "Key")
	}
	return o
}

func (o *Label) SetValue(v *string) *Label {
	if o.Value = v; o.Value == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}

// endregion

// region Taints

func (o Taint) MarshalJSON() ([]byte, error) {
	type noMethod Taint
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Taint) SetKey(v *string) *Taint {
	if o.Key = v; o.Key == nil {
		o.nullFields = append(o.nullFields, "Key")
	}
	return o
}

func (o *Taint) SetValue(v *string) *Taint {
	if o.Value = v; o.Value == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}

func (o *Taint) SetEffect(v *string) *Taint {
	if o.Effect = v; o.Effect == nil {
		o.nullFields = append(o.nullFields, "Effect")
	}
	return o
}

// endregion

// region Import

type ImportOceanGKELaunchSpecInput struct {
	OceanId      *string `json:"oceanId,omitempty"`
	NodePoolName *string `json:"nodePoolName,omitempty"`
}

// TODO: Might use LaunchSpec directly
type ImportOceanGKELaunchSpecOutput struct {
	LaunchSpec *LaunchSpec `json:"launchSpec,omitempty"`
}

// endregion

// region AutoScale

func (o AutoScale) MarshalJSON() ([]byte, error) {
	type noMethod AutoScale
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *AutoScale) SetHeadrooms(v []*AutoScaleHeadroom) *AutoScale {
	if o.Headrooms = v; o.Headrooms == nil {
		o.nullFields = append(o.nullFields, "Headrooms")
	}
	return o
}

func (o *AutoScale) SetAutoHeadroomPercentage(v *int) *AutoScale {
	if o.AutoHeadroomPercentage = v; o.AutoHeadroomPercentage == nil {
		o.nullFields = append(o.nullFields, "AutoHeadroomPercentage")
	}
	return o
}

//endregion

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

func (o *AutoScaleHeadroom) SetGPUPerUnit(v *int) *AutoScaleHeadroom {
	if o.GPUPerUnit = v; o.GPUPerUnit == nil {
		o.nullFields = append(o.nullFields, "GPUPerUnit")
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

// region Strategy

func (o LaunchSpecStrategy) MarshalJSON() ([]byte, error) {
	type noMethod LaunchSpecStrategy
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *LaunchSpecStrategy) SetPreemptiblePercentage(v *int) *LaunchSpecStrategy {
	if o.PreemptiblePercentage = v; o.PreemptiblePercentage == nil {
		o.nullFields = append(o.nullFields, "PreemptiblePercentage")
	}
	return o
}

//endregion

// region ShieldedInstanceConfig

func (o ShieldedInstanceConfig) MarshalJSON() ([]byte, error) {
	type noMethod ShieldedInstanceConfig
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ShieldedInstanceConfig) SetEnableIntegrityMonitoring(v *bool) *ShieldedInstanceConfig {
	if o.EnableIntegrityMonitoring = v; o.EnableIntegrityMonitoring == nil {
		o.nullFields = append(o.nullFields, "EnableIntegrityMonitoring")
	}
	return o
}

func (o *ShieldedInstanceConfig) SetEnableSecureBoot(v *bool) *ShieldedInstanceConfig {
	if o.EnableSecureBoot = v; o.EnableSecureBoot == nil {
		o.nullFields = append(o.nullFields, "EnableSecureBoot")
	}
	return o
}

//endregion

// region Storage

func (o Storage) MarshalJSON() ([]byte, error) {
	type noMethod Storage
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Storage) SetLocalSSDCount(v *int) *Storage {
	if o.LocalSSDCount = v; o.LocalSSDCount == nil {
		o.nullFields = append(o.nullFields, "LocalSSDCount")
	}
	return o
}

//endregion

// region ResourceLimits

func (o ResourceLimits) MarshalJSON() ([]byte, error) {
	type noMethod ResourceLimits
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ResourceLimits) SetMaxInstanceCount(v *int) *ResourceLimits {
	if o.MaxInstanceCount = v; o.MaxInstanceCount == nil {
		o.nullFields = append(o.nullFields, "MaxInstanceCount")
	}
	return o
}

func (o *ResourceLimits) SetMinInstanceCount(v *int) *ResourceLimits {
	if o.MinInstanceCount = v; o.MinInstanceCount == nil {
		o.nullFields = append(o.nullFields, "MinInstanceCount")
	}
	return o
}

//endregion

//region Scheduling

func (o GKELaunchSpecScheduling) MarshalJSON() ([]byte, error) {
	type noMethod GKELaunchSpecScheduling
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *GKELaunchSpecScheduling) SetTasks(v []*GKELaunchSpecTask) *GKELaunchSpecScheduling {
	if o.Tasks = v; o.Tasks == nil {
		o.nullFields = append(o.nullFields, "Tasks")
	}
	return o
}

// endregion

//region LaunchSpecTask

func (o GKELaunchSpecTask) MarshalJSON() ([]byte, error) {
	type noMethod GKELaunchSpecTask
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *GKELaunchSpecTask) SetIsEnabled(v *bool) *GKELaunchSpecTask {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}

func (o *GKELaunchSpecTask) SetCronExpression(v *string) *GKELaunchSpecTask {
	if o.CronExpression = v; o.CronExpression == nil {
		o.nullFields = append(o.nullFields, "CronExpression")
	}
	return o
}

func (o *GKELaunchSpecTask) SetTaskType(v *string) *GKELaunchSpecTask {
	if o.TaskType = v; o.TaskType == nil {
		o.nullFields = append(o.nullFields, "TaskType")
	}
	return o
}

func (o *GKELaunchSpecTask) SetTaskConfig(v *GKETaskConfig) *GKELaunchSpecTask {
	if o.Config = v; o.Config == nil {
		o.nullFields = append(o.nullFields, "Config")
	}
	return o
}

// endregion

//region TaskConfig

func (o GKETaskConfig) MarshalJSON() ([]byte, error) {
	type noMethod GKETaskConfig
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *GKETaskConfig) SetHeadrooms(v []*GKELaunchSpecTaskHeadroom) *GKETaskConfig {
	if o.TaskHeadrooms = v; o.TaskHeadrooms == nil {
		o.nullFields = append(o.nullFields, "TaskHeadroom")
	}
	return o
}

// endregion

// region LaunchSpecTaskHeadroom

func (o GKELaunchSpecTaskHeadroom) MarshalJSON() ([]byte, error) {
	type noMethod GKELaunchSpecTaskHeadroom
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *GKELaunchSpecTaskHeadroom) SetCPUPerUnit(v *int) *GKELaunchSpecTaskHeadroom {
	if o.CPUPerUnit = v; o.CPUPerUnit == nil {
		o.nullFields = append(o.nullFields, "CPUPerUnit")
	}
	return o
}

func (o *GKELaunchSpecTaskHeadroom) SetGPUPerUnit(v *int) *GKELaunchSpecTaskHeadroom {
	if o.GPUPerUnit = v; o.GPUPerUnit == nil {
		o.nullFields = append(o.nullFields, "GPUPerUnit")
	}
	return o
}

func (o *GKELaunchSpecTaskHeadroom) SetMemoryPerUnit(v *int) *GKELaunchSpecTaskHeadroom {
	if o.MemoryPerUnit = v; o.MemoryPerUnit == nil {
		o.nullFields = append(o.nullFields, "MemoryPerUnit")
	}
	return o
}

func (o *GKELaunchSpecTaskHeadroom) SetNumOfUnits(v *int) *GKELaunchSpecTaskHeadroom {
	if o.NumOfUnits = v; o.NumOfUnits == nil {
		o.nullFields = append(o.nullFields, "NumOfUnits")
	}
	return o
}

// endregion
