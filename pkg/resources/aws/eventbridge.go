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

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"k8s.io/klog/v2"

	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func DumpEventBridgeRule(op *resources.DumpOperation, r *resources.Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["name"] = r.Name
	data["type"] = r.Type
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)

	return nil
}

func EventBridgeRuleDeleter(cloud fi.Cloud, r *resources.Resource) error {
	return DeleteEventBridgeRule(cloud, r.Name)
}

func DeleteEventBridgeRule(cloud fi.Cloud, ruleName string) error {

	c := cloud.(awsup.AWSCloud)

	targets, err := c.EventBridge().ListTargetsByRule(&eventbridge.ListTargetsByRuleInput{
		Rule: aws.String(ruleName),
	})
	if err != nil {
		return fmt.Errorf("listing targets for EventBridge rule %q: %w", ruleName, err)
	}
	if len(targets.Targets) > 0 {
		var ids []*string
		for _, target := range targets.Targets {
			ids = append(ids, target.Id)
		}
		klog.V(2).Infof("Removing EventBridge Targets for rule %q", ruleName)
		_, err = c.EventBridge().RemoveTargets(&eventbridge.RemoveTargetsInput{
			Ids:  ids,
			Rule: aws.String(ruleName),
		})
		if err != nil {
			return fmt.Errorf("removing targets for EventBridge rule %q: %w", ruleName, err)
		}
	}

	klog.V(2).Infof("Deleting EventBridge rule %q", ruleName)
	request := &eventbridge.DeleteRuleInput{
		Name: aws.String(ruleName),
	}
	_, err = c.EventBridge().DeleteRule(request)
	if err != nil {
		return fmt.Errorf("deleting EventBridge rule %q: %w", ruleName, err)
	}
	return nil
}

func ListEventBridgeRules(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(awsup.AWSCloud)

	klog.V(2).Infof("Listing EventBridge rules")
	clusterNamePrefix := awsup.GetClusterName40(clusterName)

	// rule names start with the cluster name so that we can search for them
	request := &eventbridge.ListRulesInput{
		EventBusName: nil,
		Limit:        nil,
		NamePrefix:   aws.String(clusterNamePrefix),
	}
	response, err := c.EventBridge().ListRules(request)
	if err != nil {
		return nil, fmt.Errorf("error listing Eventbridge rules: %v", err)
	}
	if response == nil || len(response.Rules) == 0 {
		return nil, nil
	}

	var resourceTrackers []*resources.Resource

	for _, rule := range response.Rules {
		resourceTracker := &resources.Resource{
			Name:    *rule.Name,
			ID:      *rule.Name,
			Type:    "eventbridge",
			Deleter: EventBridgeRuleDeleter,
			Dumper:  DumpEventBridgeRule,
			Obj:     rule,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}
