package azure_np

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
	ID                 *string             `json:"id,omitempty"`
	OceanID            *string             `json:"oceanId,omitempty"`
	Name               *string             `json:"name,omitempty"`
	Labels             *map[string]string  `json:"labels,omitempty"`
	AvailabilityZones  []string            `json:"availabilityZones,omitempty"`
	Tags               *map[string]string  `json:"tags,omitempty"`
	Strategy           *Strategy           `json:"strategy,omitempty"`
	NodePoolProperties *NodePoolProperties `json:"nodePoolProperties,omitempty"`
	NodeCountLimits    *NodeCountLimits    `json:"nodeCountLimits,omitempty"`
	AutoScale          *AutoScale          `json:"autoScale,omitempty"`
	Taints             []*Taint            `json:"taints,omitempty"`
	VmSizes            *VmSizes            `json:"vmSizes,omitempty"`

	// Read-only fields.
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`

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
	r := client.NewRequest(http.MethodGet, "/ocean/azure/np/virtualNodeGroup")

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
	r := client.NewRequest(http.MethodPost, "/ocean/azure/np/virtualNodeGroup")
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
	path, err := uritemplates.Expand("/ocean/azure/np/virtualNodeGroup/{virtualNodeGroupId}", uritemplates.Values{
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
	path, err := uritemplates.Expand("/ocean/azure/np/virtualNodeGroup/{virtualNodeGroupId}", uritemplates.Values{
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
	path, err := uritemplates.Expand("/ocean/azure/np/virtualNodeGroup/{virtualNodeGroupId}", uritemplates.Values{
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

func (o *VirtualNodeGroup) SetLabels(v *map[string]string) *VirtualNodeGroup {
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

func (o *VirtualNodeGroup) SetAvailabilityZones(v []string) *VirtualNodeGroup {
	if o.AvailabilityZones = v; o.AvailabilityZones == nil {
		o.nullFields = append(o.nullFields, "AvailabilityZones")
	}
	return o
}

func (o *VirtualNodeGroup) SetStrategy(v *Strategy) *VirtualNodeGroup {
	if o.Strategy = v; o.Strategy == nil {
		o.nullFields = append(o.nullFields, "Strategy")
	}
	return o
}

func (o *VirtualNodeGroup) SetNodePoolProperties(v *NodePoolProperties) *VirtualNodeGroup {
	if o.NodePoolProperties = v; o.NodePoolProperties == nil {
		o.nullFields = append(o.nullFields, "NodePoolProperties")
	}
	return o
}

func (o *VirtualNodeGroup) SetNodeCountLimits(v *NodeCountLimits) *VirtualNodeGroup {
	if o.NodeCountLimits = v; o.NodeCountLimits == nil {
		o.nullFields = append(o.nullFields, "NodeCountLimits")
	}
	return o
}

func (o *VirtualNodeGroup) SetTags(v *map[string]string) *VirtualNodeGroup {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

func (o *VirtualNodeGroup) SetAutoScale(v *AutoScale) *VirtualNodeGroup {
	if o.AutoScale = v; o.AutoScale == nil {
		o.nullFields = append(o.nullFields, "AutoScale")
	}
	return o
}

func (o *VirtualNodeGroup) SetVmSizes(v *VmSizes) *VirtualNodeGroup {
	if o.VmSizes = v; o.VmSizes == nil {
		o.nullFields = append(o.nullFields, "VmSizes")
	}
	return o
}

//endregion
