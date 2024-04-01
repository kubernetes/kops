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
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"k8s.io/klog/v2"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
)

// Compile time check for interface adherence
var _ dnsprovider.ResourceRecordChangeset = &ResourceRecordChangeset{}

type ResourceRecordChangeset struct {
	zone   *Zone
	rrsets *ResourceRecordSets

	additions []dnsprovider.ResourceRecordSet
	removals  []dnsprovider.ResourceRecordSet
	upserts   []dnsprovider.ResourceRecordSet
}

func (c *ResourceRecordChangeset) Add(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	c.additions = append(c.additions, rrset)
	return c
}

func (c *ResourceRecordChangeset) Remove(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	c.removals = append(c.removals, rrset)
	return c
}

func (c *ResourceRecordChangeset) Upsert(rrset dnsprovider.ResourceRecordSet) dnsprovider.ResourceRecordChangeset {
	c.upserts = append(c.upserts, rrset)
	return c
}

// buildChange converts a dnsprovider.ResourceRecordSet to a route53.Change request
func buildChange(action route53types.ChangeAction, rrs dnsprovider.ResourceRecordSet) route53types.Change {
	change := route53types.Change{
		Action: action,
		ResourceRecordSet: &route53types.ResourceRecordSet{
			Name: aws.String(rrs.Name()),
			Type: route53types.RRType(rrs.Type()),
			TTL:  aws.Int64(rrs.Ttl()),
		},
	}

	for _, rrdata := range rrs.Rrdatas() {
		rr := route53types.ResourceRecord{
			Value: aws.String(rrdata),
		}
		change.ResourceRecordSet.ResourceRecords = append(change.ResourceRecordSet.ResourceRecords, rr)
	}
	return change
}

func (c *ResourceRecordChangeset) Apply(ctx context.Context) error {
	// Empty changesets should be a relatively quick no-op
	if c.IsEmpty() {
		return nil
	}

	hostedZoneID := c.zone.impl.Id

	removals := make(map[string]route53types.Change)
	for _, removal := range c.removals {
		removals[string(removal.Type())+"::"+removal.Name()] = buildChange(route53types.ChangeActionDelete, removal)
	}

	additions := make(map[string]route53types.Change)
	for _, addition := range c.additions {
		additions[string(addition.Type())+"::"+addition.Name()] = buildChange(route53types.ChangeActionCreate, addition)
	}

	upserts := make(map[string]route53types.Change)
	for _, upsert := range c.upserts {
		upserts[string(upsert.Type())+"::"+upsert.Name()] = buildChange(route53types.ChangeActionUpsert, upsert)
	}

	doneKeys := make(map[string]bool)

	keys := make(map[string]bool)
	for k := range removals {
		keys[k] = true
	}
	for k := range additions {
		keys[k] = true
	}
	for k := range upserts {
		keys[k] = true
	}

	for {
		var batch []route53types.Change
		// We group the changes so that changes with the same key are in the same batch
		for k := range keys {
			if doneKeys[k] {
				continue
			}

			if len(batch)+3 >= MaxBatchSize {
				break
			}

			if change, ok := removals[k]; ok {
				batch = append(batch, change)
			}
			if change, ok := additions[k]; ok {
				batch = append(batch, change)
			}
			if change, ok := upserts[k]; ok {
				batch = append(batch, change)
			}
			doneKeys[k] = true
		}

		if len(batch) == 0 {
			// Nothing left to do
			break
		}

		if klog.V(8).Enabled() {
			var sb bytes.Buffer
			for _, change := range batch {
				sb.WriteString(fmt.Sprintf("\t%s %s %s\n", change.Action, change.ResourceRecordSet.Type, aws.ToString(change.ResourceRecordSet.Name)))
			}

			klog.V(8).Infof("Route53 MaxBatchSize: %v\n", MaxBatchSize)
			klog.V(8).Infof("Route53 Changeset:\n%s", sb.String())
		}

		service := c.zone.zones.interface_.service

		request := &route53.ChangeResourceRecordSetsInput{
			ChangeBatch: &route53types.ChangeBatch{
				Changes: batch,
			},
			HostedZoneId: hostedZoneID,
		}

		// The aws-sdk-go does backoff for PriorRequestNotComplete
		_, err := service.ChangeResourceRecordSets(ctx, request)
		if err != nil {
			// Cast err to awserr.Error to get the Code and
			// Message from an error.
			return err
		}
	}

	return nil
}

func (c *ResourceRecordChangeset) IsEmpty() bool {
	return len(c.removals) == 0 && len(c.additions) == 0 && len(c.upserts) == 0
}

// ResourceRecordSets returns the parent ResourceRecordSets
func (c *ResourceRecordChangeset) ResourceRecordSets() dnsprovider.ResourceRecordSets {
	return c.rrsets
}
