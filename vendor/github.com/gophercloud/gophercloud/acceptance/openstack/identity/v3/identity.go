package v3

import (
	"testing"

	"github.com/gophercloud/gophercloud/openstack/identity/v3/endpoints"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/services"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
)

// PrintEndpoint will print an endpoint and all of its attributes.
func PrintEndpoint(t *testing.T, endpoint *endpoints.Endpoint) {
	t.Logf("ID: %s", endpoint.ID)
	t.Logf("Availability: %s", endpoint.Availability)
	t.Logf("Name: %s", endpoint.Name)
	t.Logf("Region: %s", endpoint.Region)
	t.Logf("ServiceID: %s", endpoint.ServiceID)
	t.Logf("URL: %s", endpoint.URL)
}

// PrintProject will print a project and all of its attributes.
func PrintProject(t *testing.T, project *projects.Project) {
	t.Logf("ID: %s", project.ID)
	t.Logf("IsDomain: %t", project.IsDomain)
	t.Logf("Description: %s", project.Description)
	t.Logf("DomainID: %s", project.DomainID)
	t.Logf("Enabled: %t", project.Enabled)
	t.Logf("Name: %s", project.Name)
	t.Logf("ParentID: %s", project.ParentID)
}

// PrintService will print a service and all of its attributes.
func PrintService(t *testing.T, service *services.Service) {
	t.Logf("ID: %s", service.ID)
	t.Logf("Description: %s", service.Description)
	t.Logf("Name: %s", service.Name)
	t.Logf("Type: %s", service.Type)
}

// PrintToken will print a token and all of its attributes.
func PrintToken(t *testing.T, token *tokens.Token) {
	t.Logf("ID: %s", token.ID)
	t.Logf("ExpiresAt: %v", token.ExpiresAt)
}
