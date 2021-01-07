package aws

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/uritemplates"
)

// ResourceSuggestion represents a single resource suggestion.
type ResourceSuggestion struct {
	DeploymentName  *string  `json:"deploymentName,omitempty"`
	Namespace       *string  `json:"namespace,omitempty"`
	SuggestedCPU    *float64 `json:"suggestedCPU,omitempty"`
	RequestedCPU    *float64 `json:"requestedCPU,omitempty"`
	SuggestedMemory *float64 `json:"suggestedMemory,omitempty"`
	RequestedMemory *float64 `json:"requestedMemory,omitempty"`
}

// ListResourceSuggestionsInput represents the input of `ListResourceSuggestions` function.
type ListResourceSuggestionsInput struct {
	OceanID   *string `json:"oceanId,omitempty"`
	Namespace *string `json:"namespace,omitempty"`
}

// ListResourceSuggestionsOutput represents the output of `ListResourceSuggestions` function.
type ListResourceSuggestionsOutput struct {
	Suggestions []*ResourceSuggestion `json:"suggestions,omitempty"`
}

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

// ListResourceSuggestions returns a list of right-sizing resource suggestions
// for an Ocean cluster.
func (s *ServiceOp) ListResourceSuggestions(ctx context.Context, input *ListResourceSuggestionsInput) (*ListResourceSuggestionsOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/k8s/cluster/{oceanId}/rightSizing/resourceSuggestion", uritemplates.Values{
		"oceanId": spotinst.StringValue(input.OceanID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodGet, path)

	if input.Namespace != nil {
		r.Params.Set("namespace", *input.Namespace)
	}
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := resourceSuggestionsFromHTTPResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListResourceSuggestionsOutput{Suggestions: gs}, nil
}
