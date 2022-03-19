package azure

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/uritemplates"
)

type Filter struct {
	Attribute  *Attribute `json:"attribute,omitempty"`
	Namespaces []string   `json:"namespaces,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Attribute struct {
	Key      *string `json:"key,omitempty"`
	Operator *string `json:"operator,omitempty"`
	Type     *string `json:"type,omitempty"`
	Value    *string `json:"value,omitempty"`

	forceSendFields []string
	nullFields      []string
}

// ResourceSuggestion represents a single resource suggestion.
type ResourceSuggestion struct {
	ResourceName    *string                        `json:"resourceName,omitempty"`
	ResourceType    *string                        `json:"resourceType,omitempty"`
	Namespace       *string                        `json:"namespace,omitempty"`
	SuggestedCPU    *float64                       `json:"suggestedCPU,omitempty"`
	RequestedCPU    *float64                       `json:"requestedCPU,omitempty"`
	SuggestedMemory *float64                       `json:"suggestedMemory,omitempty"`
	RequestedMemory *float64                       `json:"requestedMemory,omitempty"`
	Containers      []*ContainerResourceSuggestion `json:"containers,omitempty"`
}

// ContainerResourceSuggestion represents a resource suggestion for a
// single container.
type ContainerResourceSuggestion struct {
	Name            *string  `json:"name,omitempty"`
	SuggestedCPU    *float64 `json:"suggestedCpu,omitempty"`
	RequestedCPU    *float64 `json:"requestedCpu,omitempty"`
	SuggestedMemory *float64 `json:"suggestedMemory,omitempty"`
	RequestedMemory *float64 `json:"requestedMemory,omitempty"`
}

// ListResourceSuggestionsInput represents the input of `ListResourceSuggestions` function.
type ListResourceSuggestionsInput struct {
	OceanID   *string `json:"oceanId,omitempty"`
	Namespace *string `json:"namespace,omitempty"`
	Filter    *Filter `json:"filter,omitempty"`
}

// ListResourceSuggestionsOutput represents the output of `ListResourceSuggestions` function.
type ListResourceSuggestionsOutput struct {
	Suggestions []*ResourceSuggestion `json:"suggestions,omitempty"`
}

// region Unmarshallers

func resourceSuggestionFromJSON(in []byte) (*ResourceSuggestion, error) {
	b := new(ResourceSuggestion)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func resourceSuggestionsFromJSON(in []byte) ([]*ResourceSuggestion, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*ResourceSuggestion, len(rw.Response.Items))
	for i, rb := range rw.Response.Items {
		b, err := resourceSuggestionFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func resourceSuggestionsFromHTTPResponse(resp *http.Response) ([]*ResourceSuggestion, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return resourceSuggestionsFromJSON(body)
}

// endregion

// region API request

// ListResourceSuggestions returns a list of right-sizing resource suggestions
// for an Ocean cluster.
func (s *ServiceOp) ListResourceSuggestions(ctx context.Context, input *ListResourceSuggestionsInput) (*ListResourceSuggestionsOutput, error) {
	path, err := uritemplates.Expand("/ocean/azure/k8s/cluster/{oceanId}/rightSizing/suggestion", uritemplates.Values{
		"oceanId": spotinst.StringValue(input.OceanID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodPost, path)

	// We do NOT need the ID anymore, so let's drop it.
	input.OceanID = nil
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rs, err := resourceSuggestionsFromHTTPResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListResourceSuggestionsOutput{Suggestions: rs}, nil
}

//endregion
