package azure

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudup/azure/azuredns"
	"fmt"
)

type AzureCloud struct {


}

func (a *AzureCloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return &fi.VPCInfo{}, nil
}

func (a *AzureCloud) ProviderID() fi.CloudProviderID {
	var id fi.CloudProviderID
	id = ""
	return id
}

// DNS returns the Azure dns provider
// Because there isn't an "official" dns provider for Azure. Or at least I couldn't find one
// in the ~30 seconds or so of Googling I did. I wrote my own. It lives in kops, and doesn't
// work with GetDnsProvider() (because I am not changing k8s code for this)
func (a *AzureCloud) DNS() (dnsprovider.Interface, error) {
	//provider, err := dnsprovider.GetDnsProvider(azuredns.ProviderName, nil)
	//if err != nil {
	//	return nil, fmt.Errorf("Error building (k8s) DNS provider: %v", err)
	//}
	provider, err := azuredns.NewAzureDns()
	if err != nil {
		return nil, fmt.Errorf("error building azuredns provider: %v", err)
	}
	return provider, nil
}

func NewAzureCloud(cluster *api.Cluster) (*AzureCloud, error) {
	return &AzureCloud{}, nil
}
