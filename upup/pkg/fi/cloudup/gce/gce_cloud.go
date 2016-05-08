package gce

import (
	"fmt"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/storage/v1"
	"k8s.io/kube-deploy/upup/pkg/fi"
)

type GCECloud struct {
	Compute *compute.Service
	Storage *storage.Service

	Region  string
	Project string

	//tags    map[string]string
}

var _ fi.Cloud = &GCECloud{}

func (c *GCECloud) ProviderID() fi.CloudProviderID {
	return fi.CloudProviderGCE
}

func NewGCECloud(region string, project string) (*GCECloud, error) {
	c := &GCECloud{Region: region, Project: project}

	ctx := context.Background()

	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		return nil, fmt.Errorf("error building google API client: %v", err)
	}
	computeService, err := compute.New(client)
	if err != nil {
		return nil, fmt.Errorf("error building compute API client: %v", err)
	}
	c.Compute = computeService

	storageService, err := storage.New(client)
	if err != nil {
		return nil, fmt.Errorf("error building storage API client: %v", err)
	}
	c.Storage = storageService

	return c, nil
}
