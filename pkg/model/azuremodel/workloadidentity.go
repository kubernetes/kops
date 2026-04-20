/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package azuremodel

import (
	"fmt"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
)

// Azure built-in role definition IDs.
// See: https://learn.microsoft.com/azure/role-based-access-control/built-in-roles
const (
	// azureContributorRoleDefID is the ID of the built-in "Contributor" role.
	azureContributorRoleDefID = "b24988ac-6180-42a0-ab88-20f7382dd24c"
)

// WorkloadIdentityModelBuilder configures Azure Workload Identity resources
// (UAMI, federated identity credentials, and role assignments).
type WorkloadIdentityModelBuilder struct {
	*AzureModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &WorkloadIdentityModelBuilder{}

func (b *WorkloadIdentityModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	if !b.UseServiceAccountExternalPermissions() {
		return nil
	}

	issuerURL := fi.ValueOf(b.Cluster.Spec.KubeAPIServer.ServiceAccountIssuer)
	if issuerURL == "" {
		return fmt.Errorf("serviceAccountIssuer must be set for Azure Workload Identity")
	}

	rgName := b.Cluster.AzureResourceGroupName()
	subscriptionID := b.Cluster.Spec.CloudProvider.Azure.SubscriptionID
	identityName := b.Cluster.AzureWorkloadIdentityName()

	// UAMI task — idempotent, may already exist from pre-flight in apply_cluster.go.
	uami := &azuretasks.ManagedIdentity{
		Name:          fi.PtrTo(identityName),
		Lifecycle:     b.Lifecycle,
		ResourceGroup: b.LinkToResourceGroup(),
		Tags:          map[string]*string{},
	}
	c.AddTask(uami)

	// Role assignment: Contributor on the resource group.
	resourceGroupScope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionID, rgName)
	c.AddTask(&azuretasks.RoleAssignment{
		Name:            fi.PtrTo("wi-uami-contributor"),
		Lifecycle:       b.Lifecycle,
		Scope:           fi.PtrTo(resourceGroupScope),
		ManagedIdentity: uami,
		RoleDefID:       fi.PtrTo(azureContributorRoleDefID),
	})

	// Federated Identity Credentials — one per service account.
	saBindings := []struct {
		name      string
		namespace string
		sa        string
	}{
		{name: "fic-ccm", namespace: "kube-system", sa: "cloud-controller-manager"},
		{name: "fic-csi-azuredisk", namespace: "kube-system", sa: "csi-azuredisk-controller-sa"},
	}

	for _, binding := range saBindings {
		c.AddTask(&azuretasks.FederatedIdentityCredential{
			Name:            fi.PtrTo(binding.name),
			Lifecycle:       b.Lifecycle,
			ManagedIdentity: uami,
			ResourceGroup:   b.LinkToResourceGroup(),
			Issuer:          fi.PtrTo(issuerURL),
			Subject:         fi.PtrTo(fmt.Sprintf("system:serviceaccount:%s:%s", binding.namespace, binding.sa)),
			Audiences:       []*string{fi.PtrTo("api://AzureADTokenExchange")},
		})
	}

	return nil
}
