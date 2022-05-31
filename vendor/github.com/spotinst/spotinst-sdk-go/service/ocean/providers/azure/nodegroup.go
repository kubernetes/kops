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

type VirtualNodeGroup struct {
	ID                  *string                              `json:"id,omitempty"`
	OceanID             *string                              `json:"oceanId,omitempty"`
	Name                *string                              `json:"name,omitempty"`
	Labels              []*Label                             `json:"labels,omitempty"`
	Taints              []*Taint                             `json:"taints,omitempty"`
	AutoScale           *VirtualNodeGroupAutoScale           `json:"autoScale,omitempty"`
	ResourceLimits      *VirtualNodeGroupResourceLimits      `json:"resourceLimits,omitempty"`
	LaunchSpecification *VirtualNodeGroupLaunchSpecification `json:"launchSpecification,omitempty"`
	Zones               []string                             `json:"zones,omitempty"`

	// Read-only fields.
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type VirtualNodeGroupResourceLimits struct {
	MaxInstanceCount *int `json:"maxInstanceCount,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type VirtualNodeGroupAutoScale struct {
	Headrooms              []*VirtualNodeGroupHeadroom `json:"headrooms,omitempty"`
	AutoHeadroomPercentage *int                        `json:"autoHeadroomPercentage,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type VirtualNodeGroupHeadroom struct {
	CPUPerUnit    *int `json:"cpuPerUnit,omitempty"`
	GPUPerUnit    *int `json:"gpuPerUnit,omitempty"`
	MemoryPerUnit *int `json:"memoryPerUnit,omitempty"`
	NumOfUnits    *int `json:"numOfUnits,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type VirtualNodeGroupLaunchSpecification struct {
	OSDisk *OSDisk `json:"osDisk,omitempty"`
	Tags   []*Tag  `json:"tags,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ListVirtualNodeGroupsInput struct {
	OceanID *string `json:"oceanId,omitempty"`
}

type ListVirtualNodeGroupsOutput struct {
	VirtualNodeGroups []*VirtualNodeGroup `json:"virtualNodeGroups,omitempty"`
}

type CreateVirtualNodeGroupInput struct {
	VirtualNodeGroup *VirtualNodeGroup `json:"virtualNodeGroup,omitempty"`
}

type CreateVirtualNodeGroupOutput struct {
	VirtualNodeGroup *VirtualNodeGroup `json:"virtualNodeGroup,omitempty"`
}

type ReadVirtualNodeGroupInput struct {
	VirtualNodeGroupID *string `json:"virtualNodeGroupId,omitempty"`
}

type ReadVirtualNodeGroupOutput struct {
	VirtualNodeGroup *VirtualNodeGroup `json:"virtualNodeGroup,omitempty"`
}

type UpdateVirtualNodeGroupInput struct {
	VirtualNodeGroup *VirtualNodeGroup `json:"virtualNodeGroup,omitempty"`
}

type UpdateVirtualNodeGroupOutput struct {
	VirtualNodeGroup *VirtualNodeGroup `json:"virtualNodeGroup,omitempty"`
}

type DeleteVirtualNodeGroupInput struct {
	VirtualNodeGroupID *string `json:"virtualNodeGroupId,omitempty"`
}

type DeleteVirtualNodeGroupOutput struct{}

func virtualNodeGroupFromJSON(in []byte) (*VirtualNodeGroup, error) {
	b := new(VirtualNodeGroup)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func virtualNodeGroupsFromJSON(in []byte) ([]*VirtualNodeGroup, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*VirtualNodeGroup, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := virtualNodeGroupFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func virtualNodeGroupsFromHttpResponse(resp *http.Response) ([]*VirtualNodeGroup, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return virtualNodeGroupsFromJSON(body)
}

func (s *ServiceOp) ListVirtualNodeGroups(ctx context.Context, input *ListVirtualNodeGroupsInput) (*ListVirtualNodeGroupsOutput, error) {
	r := client.NewRequest(http.MethodGet, "/ocean/azure/k8s/virtualNodeGroup")

	if input.OceanID != nil {
		r.Params.Set("oceanId", spotinst.StringValue(input.OceanID))
	}

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	vngs, err := virtualNodeGroupsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListVirtualNodeGroupsOutput{VirtualNodeGroups: vngs}, nil
}

func (s *ServiceOp) CreateVirtualNodeGroup(ctx context.Context, input *CreateVirtualNodeGroupInput) (*CreateVirtualNodeGroupOutput, error) {
	r := client.NewRequest(http.MethodPost, "/ocean/azure/k8s/virtualNodeGroup")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	vngs, err := virtualNodeGroupsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateVirtualNodeGroupOutput)
	if len(vngs) > 0 {
		output.VirtualNodeGroup = vngs[0]
	}

	return output, nil
}

func (s *ServiceOp) ReadVirtualNodeGroup(ctx context.Context, input *ReadVirtualNodeGroupInput) (*ReadVirtualNodeGroupOutput, error) {
	path, err := uritemplates.Expand("/ocean/azure/k8s/virtualNodeGroup/{virtualNodeGroupId}", uritemplates.Values{
		"virtualNodeGroupId": spotinst.StringValue(input.VirtualNodeGroupID),
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

	vngs, err := virtualNodeGroupsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadVirtualNodeGroupOutput)
	if len(vngs) > 0 {
		output.VirtualNodeGroup = vngs[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateVirtualNodeGroup(ctx context.Context, input *UpdateVirtualNodeGroupInput) (*UpdateVirtualNodeGroupOutput, error) {
	path, err := uritemplates.Expand("/ocean/azure/k8s/virtualNodeGroup/{virtualNodeGroupId}", uritemplates.Values{
		"virtualNodeGroupId": spotinst.StringValue(input.VirtualNodeGroup.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.VirtualNodeGroup.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	vngs, err := virtualNodeGroupsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(UpdateVirtualNodeGroupOutput)
	if len(vngs) > 0 {
		output.VirtualNodeGroup = vngs[0]
	}

	return output, nil
}

func (s *ServiceOp) DeleteVirtualNodeGroup(ctx context.Context, input *DeleteVirtualNodeGroupInput) (*DeleteVirtualNodeGroupOutput, error) {
	path, err := uritemplates.Expand("/ocean/azure/k8s/virtualNodeGroup/{virtualNodeGroupId}", uritemplates.Values{
		"virtualNodeGroupId": spotinst.StringValue(input.VirtualNodeGroupID),
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

	return &DeleteVirtualNodeGroupOutput{}, nil
}

// region VirtualNodeGroup

func (o VirtualNodeGroup) MarshalJSON() ([]byte, error) {
	type noMethod VirtualNodeGroup
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *VirtualNodeGroup) SetId(v *string) *VirtualNodeGroup {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *VirtualNodeGroup) SetOceanId(v *string) *VirtualNodeGroup {
	if o.OceanID = v; o.OceanID == nil {
		o.nullFields = append(o.nullFields, "OceanID")
	}
	return o
}

func (o *VirtualNodeGroup) SetName(v *string) *VirtualNodeGroup {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *VirtualNodeGroup) SetLabels(v []*Label) *VirtualNodeGroup {
	if o.Labels = v; o.Labels == nil {
		o.nullFields = append(o.nullFields, "Labels")
	}
	return o
}

func (o *VirtualNodeGroup) SetTaints(v []*Taint) *VirtualNodeGroup {
	if o.Taints = v; o.Taints == nil {
		o.nullFields = append(o.nullFields, "Taints")
	}
	return o
}

func (o *VirtualNodeGroup) SetResourceLimits(v *VirtualNodeGroupResourceLimits) *VirtualNodeGroup {
	if o.ResourceLimits = v; o.ResourceLimits == nil {
		o.nullFields = append(o.nullFields, "ResourceLimits")
	}
	return o
}

func (o *VirtualNodeGroup) SetLaunchSpecification(v *VirtualNodeGroupLaunchSpecification) *VirtualNodeGroup {
	if o.LaunchSpecification = v; o.LaunchSpecification == nil {
		o.nullFields = append(o.nullFields, "LaunchSpecification")
	}
	return o
}

func (o *VirtualNodeGroup) SetAutoScale(v *VirtualNodeGroupAutoScale) *VirtualNodeGroup {
	if o.AutoScale = v; o.AutoScale == nil {
		o.nullFields = append(o.nullFields, "AutoScale")
	}
	return o
}

func (o *VirtualNodeGroup) SetZones(v []string) *VirtualNodeGroup {
	if o.Zones = v; o.Zones == nil {
		o.nullFields = append(o.nullFields, "Zones")
	}
	return o
}

// endregion

// region VirtualNodeGroupAutoScale

func (o VirtualNodeGroupAutoScale) MarshalJSON() ([]byte, error) {
	type noMethod VirtualNodeGroupAutoScale
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *VirtualNodeGroupAutoScale) SetHeadrooms(v []*VirtualNodeGroupHeadroom) *VirtualNodeGroupAutoScale {
	if o.Headrooms = v; o.Headrooms == nil {
		o.nullFields = append(o.nullFields, "Headrooms")
	}
	return o
}

func (o *VirtualNodeGroupAutoScale) SetAutoHeadroomPercentage(v *int) *VirtualNodeGroupAutoScale {
	if o.AutoHeadroomPercentage = v; o.AutoHeadroomPercentage == nil {
		o.nullFields = append(o.nullFields, "AutoHeadroomPercentage")
	}
	return o
}

//endregion

// region VirtualNodeGroupHeadroom

func (o VirtualNodeGroupHeadroom) MarshalJSON() ([]byte, error) {
	type noMethod VirtualNodeGroupHeadroom
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *VirtualNodeGroupHeadroom) SetCPUPerUnit(v *int) *VirtualNodeGroupHeadroom {
	if o.CPUPerUnit = v; o.CPUPerUnit == nil {
		o.nullFields = append(o.nullFields, "CPUPerUnit")
	}
	return o
}

func (o *VirtualNodeGroupHeadroom) SetGPUPerUnit(v *int) *VirtualNodeGroupHeadroom {
	if o.GPUPerUnit = v; o.GPUPerUnit == nil {
		o.nullFields = append(o.nullFields, "GPUPerUnit")
	}
	return o
}

func (o *VirtualNodeGroupHeadroom) SetMemoryPerUnit(v *int) *VirtualNodeGroupHeadroom {
	if o.MemoryPerUnit = v; o.MemoryPerUnit == nil {
		o.nullFields = append(o.nullFields, "MemoryPerUnit")
	}
	return o
}

func (o *VirtualNodeGroupHeadroom) SetNumOfUnits(v *int) *VirtualNodeGroupHeadroom {
	if o.NumOfUnits = v; o.NumOfUnits == nil {
		o.nullFields = append(o.nullFields, "NumOfUnits")
	}
	return o
}

// endregion

// region VirtualNodeGroupResourceLimits

func (o VirtualNodeGroupResourceLimits) MarshalJSON() ([]byte, error) {
	type noMethod VirtualNodeGroupResourceLimits
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *VirtualNodeGroupResourceLimits) SetMaxInstanceCount(v *int) *VirtualNodeGroupResourceLimits {
	if o.MaxInstanceCount = v; o.MaxInstanceCount == nil {
		o.nullFields = append(o.nullFields, "MaxInstanceCount")
	}
	return o
}

// endregion

// region VirtualNodeGroupLaunchSpecification

func (o VirtualNodeGroupLaunchSpecification) MarshalJSON() ([]byte, error) {
	type noMethod VirtualNodeGroupLaunchSpecification
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *VirtualNodeGroupLaunchSpecification) SetOSDisk(v *OSDisk) *VirtualNodeGroupLaunchSpecification {
	if o.OSDisk = v; o.OSDisk == nil {
		o.nullFields = append(o.nullFields, "OSDisk")
	}
	return o
}

func (o *VirtualNodeGroupLaunchSpecification) SetTags(v []*Tag) *VirtualNodeGroupLaunchSpecification {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

// endregion
