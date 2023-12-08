/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloud

import (
	"context"
	"fmt"
	"net/http"
	"time"

	alpha "google.golang.org/api/compute/v0.alpha"
	beta "google.golang.org/api/compute/v0.beta"
	compute "google.golang.org/api/compute/v1"
	ga "google.golang.org/api/compute/v1"
	"google.golang.org/api/networkservices/v1"
	networkservicesga "google.golang.org/api/networkservices/v1"
	networkservicesbeta "google.golang.org/api/networkservices/v1beta1"
	"google.golang.org/api/option"
	"k8s.io/klog/v2"
)

// Service is the top-level adapter for all of the different compute API
// versions.
type Service struct {
	GA                  *ga.Service
	Alpha               *alpha.Service
	Beta                *beta.Service
	NetworkServicesGA   *networkservicesga.ProjectsLocationsService
	NetworkServicesBeta *networkservicesbeta.ProjectsLocationsService
	ProjectRouter       ProjectRouter
	RateLimiter         RateLimiter
}

// NewService returns a new Service instance initialized with from an HTTP
// client to the API endpoints.
func NewService(ctx context.Context, client *http.Client, pr ProjectRouter, rl RateLimiter) (*Service, error) {
	alpha, err := alpha.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	beta, err := beta.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	ga, err := compute.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	nsGA, err := networkservicesga.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	nsBeta, err := networkservicesbeta.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	svc := &Service{
		GA:                  ga,
		Alpha:               alpha,
		Beta:                beta,
		NetworkServicesGA:   nsGA.Projects.Locations,
		NetworkServicesBeta: nsBeta.Projects.Locations,
		ProjectRouter:       pr,
		RateLimiter:         rl,
	}

	return svc, nil
}

// wrapOperation wraps a GCE anyOP in a version generic operation type.
func (s *Service) wrapOperation(anyOp any) (operation, error) {
	switch o := anyOp.(type) {
	case *ga.Operation:
		r, err := ParseResourceURL(o.SelfLink)
		if err != nil {
			return nil, fmt.Errorf("wrapOperation: %w", err)
		}
		return &gaOperation{
			s:         s,
			projectID: r.ProjectID,
			key:       r.Key,
		}, nil
	case *alpha.Operation:
		r, err := ParseResourceURL(o.SelfLink)
		if err != nil {
			return nil, fmt.Errorf("wrapOperation: %w", err)
		}
		return &alphaOperation{
			s:         s,
			projectID: r.ProjectID,
			key:       r.Key,
		}, nil
	case *beta.Operation:
		r, err := ParseResourceURL(o.SelfLink)
		if err != nil {
			return nil, fmt.Errorf("wrapOperation: %w", err)
		}
		return &betaOperation{
			s: s, projectID: r.ProjectID,
			key: r.Key,
		}, nil
	case *networkservices.Operation:
		result, err := parseNetworkServiceOpURL(o.Name)
		if err != nil {
			return nil, fmt.Errorf("wrapOperation: %w", err)
		}
		return &networkServicesOperation{
			s:         s,
			projectID: result.projectID,
			key:       result.key,
		}, nil
	case *networkservicesbeta.Operation:
		result, err := parseNetworkServiceOpURL(o.Name)
		if err != nil {
			return nil, fmt.Errorf("wrapOperation: %w", err)
		}
		// Reuse the GA operation stream for Beta.
		return &networkServicesOperation{
			s:         s,
			projectID: result.projectID,
			key:       result.key,
		}, nil
	default:
		return nil, fmt.Errorf("invalid type %T", anyOp)
	}
}

// WaitForCompletion of a long running operation. This will poll the state of
// GCE for the completion status of the given operation. genericOp can be one
// of alpha, beta, ga Operation types.
func (s *Service) WaitForCompletion(ctx context.Context, genericOp interface{}) error {
	op, err := s.wrapOperation(genericOp)
	if err != nil {
		klog.Errorf("wrapOperation(%+v) error: %v", genericOp, err)
		return err
	}

	return s.pollOperation(ctx, op)
}

// pollOperation calls operations.isDone until the function comes back true or context is Done.
// If an error occurs retrieving the operation, the loop will continue until the context is done.
// This is to prevent a transient error from bubbling up to controller-level logic.
func (s *Service) pollOperation(ctx context.Context, op operation) error {
	start := time.Now()
	var pollCount int
	for {
		// Check if context has been cancelled. Note that ctx.Done() must be checked before
		// returning ctx.Err().
		select {
		case <-ctx.Done():
			klog.V(5).Infof("op.pollOperation(%v, %v) not completed, poll count = %d, ctx.Err = %v (%v elapsed)", ctx, op, pollCount, ctx.Err(), time.Since(start))
			return ctx.Err()
		default:
			// ctx is not canceled, continue immediately
		}

		pollCount++
		klog.V(5).Infof("op.isDone(%v) waiting; op = %v, poll count = %d (%v elapsed)", ctx, op, pollCount, time.Since(start))
		s.RateLimiter.Accept(ctx, op.rateLimitKey())
		switch done, err := op.isDone(ctx); {
		case err != nil:
			klog.V(5).Infof("op.isDone(%v) error; op = %v, poll count = %d, err = %v, retrying (%v elapsed)", ctx, op, pollCount, err, time.Since(start))
			s.RateLimiter.Observe(ctx, err, op.rateLimitKey())
			return err
		case done:
			klog.V(5).Infof("op.isDone(%v) complete; op = %v, poll count = %d, op.err = %v (%v elapsed)", ctx, op, pollCount, op.error(), time.Since(start))
			s.RateLimiter.Observe(ctx, op.error(), op.rateLimitKey())
			return op.error()
		}
	}
}
