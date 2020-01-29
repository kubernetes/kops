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

package main

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	upgradeLong = templates.LongDesc(i18n.T(`
	Automates checking for and applying Kubernetes updates. This upgrades a cluster to the latest recommended
	production ready k8s version. After this command is run, use kops update cluster and kops rolling-update cluster
	to finish a cluster upgrade.
	`))

	upgradeExample = templates.Examples(i18n.T(`
	# Upgrade a cluster's Kubernetes version.
	kops upgrade cluster kubernetes-cluster.example.com --yes --state=s3://kops-state-1234
	`))

	upgradeShort = i18n.T("Upgrade a kubernetes cluster.")
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	Short:   upgradeShort,
	Long:    upgradeLong,
	Example: upgradeExample,
}

func init() {
	rootCommand.AddCommand(upgradeCmd)
}
