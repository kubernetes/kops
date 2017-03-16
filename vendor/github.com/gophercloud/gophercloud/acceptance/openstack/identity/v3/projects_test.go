// +build acceptance

package v3

import (
	"testing"

	"github.com/gophercloud/gophercloud/acceptance/clients"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
)

func TestProjectsList(t *testing.T) {
	client, err := clients.NewIdentityV3Client()
	if err != nil {
		t.Fatalf("Unable to obtain an identity client: %v")
	}

	var iTrue bool = true
	listOpts := projects.ListOpts{
		Enabled: &iTrue,
	}

	allPages, err := projects.List(client, listOpts).AllPages()
	if err != nil {
		t.Fatalf("Unable to list projects: %v", err)
	}

	allProjects, err := projects.ExtractProjects(allPages)
	if err != nil {
		t.Fatalf("Unable to extract projects: %v", err)
	}

	for _, project := range allProjects {
		PrintProject(t, &project)
	}
}
