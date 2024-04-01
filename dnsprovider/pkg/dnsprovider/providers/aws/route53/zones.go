/*
Copyright 2019 The Kubernetes Authors.

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

package route53

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"

	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
)

// Compile time check for interface adherence
var _ dnsprovider.Zones = Zones{}

type Zones struct {
	interface_ *Interface
}

func (zones Zones) List() ([]dnsprovider.Zone, error) {
	var zoneList []dnsprovider.Zone

	input := &route53.ListHostedZonesInput{}
	paginator := route53.NewListHostedZonesPaginator(zones.interface_.service, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return []dnsprovider.Zone{}, fmt.Errorf("error listing hosted zones: %w", err)
		}
		for _, zone := range page.HostedZones {
			zoneList = append(zoneList, &Zone{&zone, &zones})
		}
	}
	return zoneList, nil
}

func (zones Zones) Add(zone dnsprovider.Zone) (dnsprovider.Zone, error) {
	dnsName := zone.Name()
	callerReference := string(uuid.NewUUID())
	input := route53.CreateHostedZoneInput{Name: &dnsName, CallerReference: &callerReference}
	output, err := zones.interface_.service.CreateHostedZone(context.TODO(), &input)
	if err != nil {
		return nil, err
	}
	return &Zone{output.HostedZone, &zones}, nil
}

func (zones Zones) Remove(zone dnsprovider.Zone) error {
	zoneId := zone.(*Zone).impl.Id
	input := route53.DeleteHostedZoneInput{Id: zoneId}
	_, err := zones.interface_.service.DeleteHostedZone(context.TODO(), &input)
	if err != nil {
		return err
	}
	return nil
}

func (zones Zones) New(name string) (dnsprovider.Zone, error) {
	id := string(uuid.NewUUID())
	managedZone := route53types.HostedZone{Id: &id, Name: &name}
	return &Zone{&managedZone, &zones}, nil
}
