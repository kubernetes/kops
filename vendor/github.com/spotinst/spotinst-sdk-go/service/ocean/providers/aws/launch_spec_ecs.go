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

type ECSLaunchSpec struct {
	ID                  *string                  `json:"id,omitempty"`
	Name                *string                  `json:"name,omitempty"`
	OceanID             *string                  `json:"oceanId,omitempty"`
	ImageID             *string                  `json:"imageId,omitempty"`
	UserData            *string                  `json:"userData,omitempty"`
	SecurityGroupIDs    []string                 `json:"securityGroupIds,omitempty"`
	AutoScale           *ECSAutoScale            `json:"autoScale,omitempty"`
	IAMInstanceProfile  *ECSIAMInstanceProfile   `json:"iamInstanceProfile,omitempty"`
	Attributes          []*ECSAttribute          `json:"attributes,omitempty"`
	BlockDeviceMappings []*ECSBlockDeviceMapping `json:"blockDeviceMappings,omitempty"`
	Tags                []*Tag                   `json:"tags,omitempty"`
	InstanceTypes       []string                 `json:"instanceTypes,omitempty"`

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

type ECSAttribute struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSAutoScale struct {
	Headrooms []*ECSAutoScaleHeadroom `json:"headrooms,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSAutoScaleHeadroom struct {
	CPUPerUnit    *int `json:"cpuPerUnit,omitempty"`
	MemoryPerUnit *int `json:"memoryPerUnit,omitempty"`
	NumOfUnits    *int `json:"numOfUnits,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSBlockDeviceMapping struct {
	DeviceName  *string `json:"deviceName,omitempty"`
	NoDevice    *string `json:"noDevice,omitempty"`
	VirtualName *string `json:"virtualName,omitempty"`
	EBS         *ECSEBS `json:"ebs,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSEBS struct {
	DeleteOnTermination *bool                 `json:"deleteOnTermination,omitempty"`
	Encrypted           *bool                 `json:"encrypted,omitempty"`
	KMSKeyID            *string               `json:"kmsKeyId,omitempty"`
	SnapshotID          *string               `json:"snapshotId,omitempty"`
	VolumeType          *string               `json:"volumeType,omitempty"`
	IOPS                *int                  `json:"iops,omitempty"`
	VolumeSize          *int                  `json:"volumeSize,omitempty"`
	Throughput          *int                  `json:"throughput,omitempty"`
	DynamicVolumeSize   *ECSDynamicVolumeSize `json:"dynamicVolumeSize,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ECSDynamicVolumeSize struct {
	BaseSize            *int    `json:"baseSize,omitempty"`
	SizePerResourceUnit *int    `json:"sizePerResourceUnit,omitempty"`
	Resource            *string `json:"resource,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ListECSLaunchSpecsInput struct {
	OceanID *string `json:"oceanId,omitempty"`
}

type ListECSLaunchSpecsOutput struct {
	LaunchSpecs []*ECSLaunchSpec `json:"launchSpecs,omitempty"`
}

type CreateECSLaunchSpecInput struct {
	LaunchSpec *ECSLaunchSpec `json:"launchSpec,omitempty"`
}

type CreateECSLaunchSpecOutput struct {
	LaunchSpec *ECSLaunchSpec `json:"launchSpec,omitempty"`
}

type ReadECSLaunchSpecInput struct {
	LaunchSpecID *string `json:"launchSpecId,omitempty"`
}

type ReadECSLaunchSpecOutput struct {
	LaunchSpec *ECSLaunchSpec `json:"launchSpec,omitempty"`
}

type UpdateECSLaunchSpecInput struct {
	LaunchSpec *ECSLaunchSpec `json:"launchSpec,omitempty"`
}

type UpdateECSLaunchSpecOutput struct {
	LaunchSpec *ECSLaunchSpec `json:"launchSpec,omitempty"`
}

type DeleteECSLaunchSpecInput struct {
	LaunchSpecID *string `json:"launchSpecId,omitempty"`
}

type DeleteECSLaunchSpecOutput struct{}

func ecsLaunchSpecFromJSON(in []byte) (*ECSLaunchSpec, error) {
	b := new(ECSLaunchSpec)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func ecsLaunchSpecsFromJSON(in []byte) ([]*ECSLaunchSpec, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*ECSLaunchSpec, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := ecsLaunchSpecFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func ecsLaunchSpecsFromHttpResponse(resp *http.Response) ([]*ECSLaunchSpec, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ecsLaunchSpecsFromJSON(body)
}

func (s *ServiceOp) ListECSLaunchSpecs(ctx context.Context, input *ListECSLaunchSpecsInput) (*ListECSLaunchSpecsOutput, error) {
	r := client.NewRequest(http.MethodGet, "/ocean/aws/ecs/launchSpec")

	if input.OceanID != nil {
		r.Params.Set("oceanId", spotinst.StringValue(input.OceanID))
	}

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := ecsLaunchSpecsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListECSLaunchSpecsOutput{LaunchSpecs: gs}, nil
}

func (s *ServiceOp) CreateECSLaunchSpec(ctx context.Context, input *CreateECSLaunchSpecInput) (*CreateECSLaunchSpecOutput, error) {
	r := client.NewRequest(http.MethodPost, "/ocean/aws/ecs/launchSpec")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := ecsLaunchSpecsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateECSLaunchSpecOutput)
	if len(gs) > 0 {
		output.LaunchSpec = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) ReadECSLaunchSpec(ctx context.Context, input *ReadECSLaunchSpecInput) (*ReadECSLaunchSpecOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/ecs/launchSpec/{launchSpecId}", uritemplates.Values{
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

	gs, err := ecsLaunchSpecsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadECSLaunchSpecOutput)
	if len(gs) > 0 {
		output.LaunchSpec = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateECSLaunchSpec(ctx context.Context, input *UpdateECSLaunchSpecInput) (*UpdateECSLaunchSpecOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/ecs/launchSpec/{launchSpecId}", uritemplates.Values{
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

	gs, err := ecsLaunchSpecsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(UpdateECSLaunchSpecOutput)
	if len(gs) > 0 {
		output.LaunchSpec = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) DeleteECSLaunchSpec(ctx context.Context, input *DeleteECSLaunchSpecInput) (*DeleteECSLaunchSpecOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/ecs/launchSpec/{launchSpecId}", uritemplates.Values{
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

	return &DeleteECSLaunchSpecOutput{}, nil
}

// endregion

// region LaunchSpec

func (o ECSLaunchSpec) MarshalJSON() ([]byte, error) {
	type noMethod ECSLaunchSpec
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSLaunchSpec) SetId(v *string) *ECSLaunchSpec {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *ECSLaunchSpec) SetName(v *string) *ECSLaunchSpec {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *ECSLaunchSpec) SetOceanId(v *string) *ECSLaunchSpec {
	if o.OceanID = v; o.OceanID == nil {
		o.nullFields = append(o.nullFields, "OceanID")
	}
	return o
}

func (o *ECSLaunchSpec) SetImageId(v *string) *ECSLaunchSpec {
	if o.ImageID = v; o.ImageID == nil {
		o.nullFields = append(o.nullFields, "ImageID")
	}
	return o
}

func (o *ECSLaunchSpec) SetUserData(v *string) *ECSLaunchSpec {
	if o.UserData = v; o.UserData == nil {
		o.nullFields = append(o.nullFields, "UserData")
	}
	return o
}

func (o *ECSLaunchSpec) SetSecurityGroupIDs(v []string) *ECSLaunchSpec {
	if o.SecurityGroupIDs = v; o.SecurityGroupIDs == nil {
		o.nullFields = append(o.nullFields, "SecurityGroupIDs")
	}
	return o
}

func (o *ECSLaunchSpec) SetIAMInstanceProfile(v *ECSIAMInstanceProfile) *ECSLaunchSpec {
	if o.IAMInstanceProfile = v; o.IAMInstanceProfile == nil {
		o.nullFields = append(o.nullFields, "IAMInstanceProfile")
	}
	return o
}

func (o *ECSLaunchSpec) SetAttributes(v []*ECSAttribute) *ECSLaunchSpec {
	if o.Attributes = v; o.Attributes == nil {
		o.nullFields = append(o.nullFields, "Attributes")
	}
	return o
}

func (o *ECSLaunchSpec) SetAutoScale(v *ECSAutoScale) *ECSLaunchSpec {
	if o.AutoScale = v; o.AutoScale == nil {
		o.nullFields = append(o.nullFields, "AutoScale")
	}
	return o
}

func (o *ECSLaunchSpec) SetBlockDeviceMappings(v []*ECSBlockDeviceMapping) *ECSLaunchSpec {
	if o.BlockDeviceMappings = v; o.BlockDeviceMappings == nil {
		o.nullFields = append(o.nullFields, "BlockDeviceMappings")
	}
	return o
}

func (o *ECSLaunchSpec) SetTags(v []*Tag) *ECSLaunchSpec {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

func (o *ECSLaunchSpec) SetInstanceTypes(v []string) *ECSLaunchSpec {
	if o.InstanceTypes = v; o.InstanceTypes == nil {
		o.nullFields = append(o.nullFields, "InstanceTypes")
	}
	return o
}

// endregion

// region Attributes

func (o ECSAttribute) MarshalJSON() ([]byte, error) {
	type noMethod ECSAttribute
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSAttribute) SetKey(v *string) *ECSAttribute {
	if o.Key = v; o.Key == nil {
		o.nullFields = append(o.nullFields, "Key")
	}
	return o
}

func (o *ECSAttribute) SetValue(v *string) *ECSAttribute {
	if o.Value = v; o.Value == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}

// endregion

// region AutoScale

func (o ECSAutoScale) MarshalJSON() ([]byte, error) {
	type noMethod ECSAutoScale
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSAutoScale) SetHeadrooms(v []*ECSAutoScaleHeadroom) *ECSAutoScale {
	if o.Headrooms = v; o.Headrooms == nil {
		o.nullFields = append(o.nullFields, "Headrooms")
	}
	return o
}

//endregion

// region ECSAutoScaleHeadroom

func (o ECSAutoScaleHeadroom) MarshalJSON() ([]byte, error) {
	type noMethod ECSAutoScaleHeadroom
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSAutoScaleHeadroom) SetCPUPerUnit(v *int) *ECSAutoScaleHeadroom {
	if o.CPUPerUnit = v; o.CPUPerUnit == nil {
		o.nullFields = append(o.nullFields, "CPUPerUnit")
	}
	return o
}

func (o *ECSAutoScaleHeadroom) SetMemoryPerUnit(v *int) *ECSAutoScaleHeadroom {
	if o.MemoryPerUnit = v; o.MemoryPerUnit == nil {
		o.nullFields = append(o.nullFields, "MemoryPerUnit")
	}
	return o
}

func (o *ECSAutoScaleHeadroom) SetNumOfUnits(v *int) *ECSAutoScaleHeadroom {
	if o.NumOfUnits = v; o.NumOfUnits == nil {
		o.nullFields = append(o.nullFields, "NumOfUnits")
	}
	return o
}

// endregion

// region ECSBlockDeviceMapping

func (o ECSBlockDeviceMapping) MarshalJSON() ([]byte, error) {
	type noMethod ECSBlockDeviceMapping
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSBlockDeviceMapping) SetDeviceName(v *string) *ECSBlockDeviceMapping {
	if o.DeviceName = v; o.DeviceName == nil {
		o.nullFields = append(o.nullFields, "DeviceName")
	}
	return o
}

func (o *ECSBlockDeviceMapping) SetNoDevice(v *string) *ECSBlockDeviceMapping {
	if o.NoDevice = v; o.NoDevice == nil {
		o.nullFields = append(o.nullFields, "NoDevice")
	}
	return o
}

func (o *ECSBlockDeviceMapping) SetVirtualName(v *string) *ECSBlockDeviceMapping {
	if o.VirtualName = v; o.VirtualName == nil {
		o.nullFields = append(o.nullFields, "VirtualName")
	}
	return o
}

func (o *ECSBlockDeviceMapping) SetEBS(v *ECSEBS) *ECSBlockDeviceMapping {
	if o.EBS = v; o.EBS == nil {
		o.nullFields = append(o.nullFields, "EBS")
	}
	return o
}

// endregion

// region ECSEBS

func (o ECSEBS) MarshalJSON() ([]byte, error) {
	type noMethod ECSEBS
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSEBS) SetEncrypted(v *bool) *ECSEBS {
	if o.Encrypted = v; o.Encrypted == nil {
		o.nullFields = append(o.nullFields, "Encrypted")
	}
	return o
}

func (o *ECSEBS) SetIOPS(v *int) *ECSEBS {
	if o.IOPS = v; o.IOPS == nil {
		o.nullFields = append(o.nullFields, "IOPS")
	}
	return o
}

func (o *ECSEBS) SetKMSKeyId(v *string) *ECSEBS {
	if o.KMSKeyID = v; o.KMSKeyID == nil {
		o.nullFields = append(o.nullFields, "KMSKeyID")
	}
	return o
}

func (o *ECSEBS) SetSnapshotId(v *string) *ECSEBS {
	if o.SnapshotID = v; o.SnapshotID == nil {
		o.nullFields = append(o.nullFields, "SnapshotID")
	}
	return o
}

func (o *ECSEBS) SetVolumeType(v *string) *ECSEBS {
	if o.VolumeType = v; o.VolumeType == nil {
		o.nullFields = append(o.nullFields, "VolumeType")
	}
	return o
}

func (o *ECSEBS) SetDeleteOnTermination(v *bool) *ECSEBS {
	if o.DeleteOnTermination = v; o.DeleteOnTermination == nil {
		o.nullFields = append(o.nullFields, "DeleteOnTermination")
	}
	return o
}

func (o *ECSEBS) SetVolumeSize(v *int) *ECSEBS {
	if o.VolumeSize = v; o.VolumeSize == nil {
		o.nullFields = append(o.nullFields, "VolumeSize")
	}
	return o
}

func (o *ECSEBS) SetDynamicVolumeSize(v *ECSDynamicVolumeSize) *ECSEBS {
	if o.DynamicVolumeSize = v; o.DynamicVolumeSize == nil {
		o.nullFields = append(o.nullFields, "DynamicVolumeSize")
	}
	return o
}

func (o *ECSEBS) SetThroughput(v *int) *ECSEBS {
	if o.Throughput = v; o.Throughput == nil {
		o.nullFields = append(o.nullFields, "Throughput")
	}
	return o
}

// endregion

// region ECSDynamicVolumeSize

func (o ECSDynamicVolumeSize) MarshalJSON() ([]byte, error) {
	type noMethod ECSDynamicVolumeSize
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ECSDynamicVolumeSize) SetBaseSize(v *int) *ECSDynamicVolumeSize {
	if o.BaseSize = v; o.BaseSize == nil {
		o.nullFields = append(o.nullFields, "BaseSize")
	}
	return o
}

func (o *ECSDynamicVolumeSize) SetResource(v *string) *ECSDynamicVolumeSize {
	if o.Resource = v; o.Resource == nil {
		o.nullFields = append(o.nullFields, "Resource")
	}
	return o
}

func (o *ECSDynamicVolumeSize) SetSizePerResourceUnit(v *int) *ECSDynamicVolumeSize {
	if o.SizePerResourceUnit = v; o.SizePerResourceUnit == nil {
		o.nullFields = append(o.nullFields, "SizePerResourceUnit")
	}
	return o
}

// endregion
