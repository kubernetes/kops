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

package openstack

import (
	"context"
	"fmt"

	sg "github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/security/groups"
	sgr "github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/security/rules"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kops/util/pkg/vfs"
)

func (c *openstackCloud) ListSecurityGroups(opt sg.ListOpts) ([]sg.SecGroup, error) {
	return listSecurityGroups(c, opt)
}

func listSecurityGroups(c OpenstackCloud, opt sg.ListOpts) ([]sg.SecGroup, error) {
	var groups []sg.SecGroup

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		allPages, err := sg.List(c.NetworkingClient(), opt).AllPages(context.TODO())
		if err != nil {
			return false, fmt.Errorf("error listing security groups %v: %v", opt, err)
		}

		gs, err := sg.ExtractGroups(allPages)
		if err != nil {
			return false, fmt.Errorf("error extracting security groups from pages: %v", err)
		}
		groups = gs
		return true, nil
	})
	if err != nil {
		return groups, err
	} else if done {
		return groups, nil
	} else {
		return groups, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) CreateSecurityGroup(opt sg.CreateOptsBuilder) (*sg.SecGroup, error) {
	return createSecurityGroup(c, opt)
}

func createSecurityGroup(c OpenstackCloud, opt sg.CreateOptsBuilder) (*sg.SecGroup, error) {
	var group *sg.SecGroup

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		g, err := sg.Create(context.TODO(), c.NetworkingClient(), opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error creating security group %v: %v", opt, err)
		}
		group = g
		return true, nil
	})
	if err != nil {
		return group, err
	} else if done {
		return group, nil
	} else {
		return group, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) ListSecurityGroupRules(opt sgr.ListOpts) ([]sgr.SecGroupRule, error) {
	return listSecurityGroupRules(c, opt)
}

func listSecurityGroupRules(c OpenstackCloud, opt sgr.ListOpts) ([]sgr.SecGroupRule, error) {
	var rules []sgr.SecGroupRule

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		allPages, err := sgr.List(c.NetworkingClient(), opt).AllPages(context.TODO())
		if err != nil {
			return false, fmt.Errorf("error listing security group rules %v: %v", opt, err)
		}

		rs, err := sgr.ExtractRules(allPages)
		if err != nil {
			return false, fmt.Errorf("error extracting security group rules from pages: %v", err)
		}
		rules = rs
		return true, nil
	})
	if err != nil {
		return rules, err
	} else if done {
		return rules, nil
	} else {
		return rules, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) CreateSecurityGroupRule(opt sgr.CreateOptsBuilder) (*sgr.SecGroupRule, error) {
	return createSecurityGroupRule(c, opt)
}

func createSecurityGroupRule(c OpenstackCloud, opt sgr.CreateOptsBuilder) (*sgr.SecGroupRule, error) {
	var rule *sgr.SecGroupRule

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		r, err := sgr.Create(context.TODO(), c.NetworkingClient(), opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error creating security group rule %v: %v", opt, err)
		}
		rule = r
		return true, nil
	})
	if err != nil {
		return rule, err
	} else if done {
		return rule, nil
	} else {
		return rule, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) DeleteSecurityGroup(sgID string) error {
	return deleteSecurityGroup(c, sgID)
}

func deleteSecurityGroup(c OpenstackCloud, sgID string) error {
	done, err := vfs.RetryWithBackoff(deleteBackoff, func() (bool, error) {
		err := sg.Delete(context.TODO(), c.NetworkingClient(), sgID).ExtractErr()
		if err != nil && !isNotFound(err) {
			return false, fmt.Errorf("error deleting security group: %v", err)
		}
		if isNotFound(err) {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	} else if done {
		return nil
	} else {
		return wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) DeleteSecurityGroupRule(ruleID string) error {
	return deleteSecurityGroupRule(c, ruleID)
}

func deleteSecurityGroupRule(c OpenstackCloud, ruleID string) error {
	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		err := sgr.Delete(context.TODO(), c.NetworkingClient(), ruleID).ExtractErr()
		if err != nil && !isNotFound(err) {
			return false, fmt.Errorf("error deleting security group rule: %v", err)
		}
		return true, nil
	})
	if err != nil {
		return err
	} else if done {
		return nil
	} else {
		return wait.ErrWaitTimeout
	}
}
