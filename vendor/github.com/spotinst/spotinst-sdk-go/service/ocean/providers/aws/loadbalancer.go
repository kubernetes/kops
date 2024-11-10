package aws

import (
	"context"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/uritemplates"
	"net/http"
)

type LoadBalancers struct {
	Arn  *string `json:"arn,omitempty"`
	Name *string `json:"name,omitempty"`
	Type *string `json:"type,omitempty"`
}

type AttachLoadbalancerInput struct {
	LoadBalancers []*LoadBalancers `json:"loadBalancers,omitempty"`
	ID            *string          `json:"id,omitempty"`
}

type AttachLoadbalancerOutput struct{}

type DetachLoadbalancerInput struct {
	LoadBalancers []*LoadBalancers `json:"loadBalancers,omitempty"`
	ID            *string          `json:"id,omitempty"`
}

type DetachLoadbalancerOutput struct{}

func (s *ServiceOp) AttachLoadBalancer(ctx context.Context,
	input *AttachLoadbalancerInput) (*AttachLoadbalancerOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/k8s/cluster/{oceanClusterId}/loadBalancer/attach", uritemplates.Values{
		"oceanClusterId": spotinst.StringValue(input.ID),
	})
	if err != nil {
		return nil, err
	}
	input.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &AttachLoadbalancerOutput{}, nil
}

func (s *ServiceOp) DetachLoadBalancer(ctx context.Context,
	input *DetachLoadbalancerInput) (*DetachLoadbalancerOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/k8s/cluster/{oceanClusterId}/loadBalancer/detach", uritemplates.Values{
		"oceanClusterId": spotinst.StringValue(input.ID),
	})
	if err != nil {
		return nil, err
	}
	input.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DetachLoadbalancerOutput{}, nil
}
