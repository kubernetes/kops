package aws

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

type ExtendedResourceDefinition struct {
	ID      *string                `json:"id,omitempty"`
	Name    *string                `json:"name,omitempty"`
	Mapping map[string]interface{} `json:"mapping,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ListExtendedResourceDefinitionsInput struct{}

type ListExtendedResourceDefinitionsOutput struct {
	ExtendedResourceDefinitions []*ExtendedResourceDefinition `json:"extendedResourceDefinitions,omitempty"`
}

type CreateExtendedResourceDefinitionInput struct {
	ExtendedResourceDefinition *ExtendedResourceDefinition `json:"extendedResourceDefinition,omitempty"`
}

type CreateExtendedResourceDefinitionOutput struct {
	ExtendedResourceDefinition *ExtendedResourceDefinition `json:"extendedResourceDefinition,omitempty"`
}

type ReadExtendedResourceDefinitionInput struct {
	ExtendedResourceDefinitionID *string `json:"oceanExtendedResourceDefinitionId,omitempty"`
}

type ReadExtendedResourceDefinitionOutput struct {
	ExtendedResourceDefinition *ExtendedResourceDefinition `json:"extendedResourceDefinition,omitempty"`
}

type UpdateExtendedResourceDefinitionInput struct {
	ExtendedResourceDefinition *ExtendedResourceDefinition `json:"extendedResourceDefinition,omitempty"`
}

type UpdateExtendedResourceDefinitionOutput struct {
	ExtendedResourceDefinition *ExtendedResourceDefinition `json:"extendedResourceDefinition,omitempty"`
}

type DeleteExtendedResourceDefinitionInput struct {
	ExtendedResourceDefinitionID *string `json:"extendedResourceDefinitionId,omitempty"`
}

type DeleteExtendedResourceDefinitionOutput struct{}

func extendedResourceDefinitionFromJSON(in []byte) (*ExtendedResourceDefinition, error) {
	b := new(ExtendedResourceDefinition)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func extendedResourceDefinitionsFromJSON(in []byte) ([]*ExtendedResourceDefinition, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*ExtendedResourceDefinition, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := extendedResourceDefinitionFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func extendedResourceDefinitionsFromHttpResponse(resp *http.Response) ([]*ExtendedResourceDefinition, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return extendedResourceDefinitionsFromJSON(body)
}

func (s *ServiceOp) ListExtendedResourceDefinition(ctx context.Context, input *ListExtendedResourceDefinitionsInput) (*ListExtendedResourceDefinitionsOutput, error) {
	r := client.NewRequest(http.MethodGet, "/ocean/k8s/extendedResourceDefinition")
	resp, err := client.RequireOK(s.Client.Do(ctx, r))

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	erds, err := extendedResourceDefinitionsFromHttpResponse(resp)

	if err != nil {
		return nil, err
	}

	return &ListExtendedResourceDefinitionsOutput{ExtendedResourceDefinitions: erds}, nil
}

func (s *ServiceOp) CreateExtendedResourceDefinition(ctx context.Context, input *CreateExtendedResourceDefinitionInput) (*CreateExtendedResourceDefinitionOutput, error) {
	r := client.NewRequest(http.MethodPost, "/ocean/k8s/extendedResourceDefinition")
	r.Obj = input
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	erds, err := extendedResourceDefinitionsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateExtendedResourceDefinitionOutput)
	if len(erds) > 0 {
		output.ExtendedResourceDefinition = erds[0]
	}

	return output, nil
}

func (s *ServiceOp) ReadExtendedResourceDefinition(ctx context.Context, input *ReadExtendedResourceDefinitionInput) (*ReadExtendedResourceDefinitionOutput, error) {
	path, err := uritemplates.Expand("/ocean/k8s/extendedResourceDefinition/{oceanExtendedResourceDefinitionId}", uritemplates.Values{
		"oceanExtendedResourceDefinitionId": spotinst.StringValue(input.ExtendedResourceDefinitionID),
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

	erds, err := extendedResourceDefinitionsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadExtendedResourceDefinitionOutput)
	if len(erds) > 0 {
		output.ExtendedResourceDefinition = erds[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateExtendedResourceDefinition(ctx context.Context, input *UpdateExtendedResourceDefinitionInput) (*UpdateExtendedResourceDefinitionOutput, error) {
	path, err := uritemplates.Expand("/ocean/k8s/extendedResourceDefinition/{oceanExtendedResourceDefinitionId}", uritemplates.Values{
		"oceanExtendedResourceDefinitionId": spotinst.StringValue(input.ExtendedResourceDefinition.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.ExtendedResourceDefinition.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	erds, err := extendedResourceDefinitionsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(UpdateExtendedResourceDefinitionOutput)
	if len(erds) > 0 {
		output.ExtendedResourceDefinition = erds[0]
	}

	return output, nil
}

func (s *ServiceOp) DeleteExtendedResourceDefinition(ctx context.Context, input *DeleteExtendedResourceDefinitionInput) (*DeleteExtendedResourceDefinitionOutput, error) {
	path, err := uritemplates.Expand("/ocean/k8s/extendedResourceDefinition/{oceanExtendedResourceDefinitionId}", uritemplates.Values{
		"oceanExtendedResourceDefinitionId": spotinst.StringValue(input.ExtendedResourceDefinitionID),
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

	return &DeleteExtendedResourceDefinitionOutput{}, nil
}

// region ExtendedResourceDefinition

func (o ExtendedResourceDefinition) MarshalJSON() ([]byte, error) {
	type noMethod ExtendedResourceDefinition
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ExtendedResourceDefinition) SetId(v *string) *ExtendedResourceDefinition {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *ExtendedResourceDefinition) SetName(v *string) *ExtendedResourceDefinition {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *ExtendedResourceDefinition) SetMapping(v map[string]interface{}) *ExtendedResourceDefinition {
	if o.Mapping = v; o.Mapping == nil {
		o.nullFields = append(o.nullFields, "Mapping")
	}
	return o
}

// endregion
