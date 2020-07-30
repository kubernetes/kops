/*
Copyright 2020 The Kubernetes Authors.

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
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/cli"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

// Filter Flag Constants
const (
	vcpus                  = "vcpus"
	memory                 = "memory"
	vcpusToMemoryRatio     = "vcpus-to-memory-ratio"
	cpuArchitecture        = "cpu-architecture"
	gpus                   = "gpus"
	gpuMemoryTotal         = "gpu-memory-total"
	placementGroupStrategy = "placement-group-strategy"
	usageClass             = "usage-class"
	enaSupport             = "ena-support"
	burstSupport           = "burst-support"
	availabilityZones      = "zones"
	networkInterfaces      = "network-interfaces"
	networkPerformance     = "network-performance"
	allowList              = "allow-list"
	denyList               = "deny-list"
	maxResults             = "max-results"
)

// Aggregate Filter Flag Constants
const (
	instanceTypeBase = "instance-type-base"
	flexible         = "flexible"
)

// Control Flag Constants
const (
	instanceGroupCount = "instance-group-count"
	nodeCountMin       = "node-count-min"
	nodeCountMax       = "node-count-max"
	nodeVolumeSize     = "node-volume-size"
	nodeSecurityGroups = "node-security-groups"
	clusterAutoscaler  = "cluster-autoscaler"
	usageClassSpot     = "spot"
	usageClassOndemand = "on-demand"
	igName             = "instance-group-name"
	dryRun             = "dry-run"
	output             = "output"
)

const (
	nameRegex = `^[a-zA-Z0-9\-_]{1,128}$`
)

// InstanceSelectorOptions is a struct representing non-filter flags passed into instance-selector
type InstanceSelectorOptions struct {
	NodeCountMin       int32
	NodeCountMax       int32
	NodeVolumeSize     *int32
	NodeSecurityGroups []string
	ClusterAutoscaler  bool
	InstanceGroupName  string
	InstanceGroupCount int
	Output             string
	DryRun             bool
}

var (
	toolboxInstanceSelectorLong = templates.LongDesc(i18n.T(`
	Generate AWS EC2 on-demand or spot instance-groups by providing resource specs like vcpus and memory rather than instance types.`))

	toolboxInstanceSelectorExample = templates.Examples(i18n.T(`

	## Create a best-practices spot instance-group using a MixInstancesPolicy and Capacity-Optimized spot allocation strategy
	## --flexible defaults to a 1:2 vcpus to memory ratio and 4 vcpus
	kops toolbox instance-selector --usage-class spot --flexible --instance-group-name my-spot-mig

	## Create a best-practices on-demand instance-group with custom vcpus and memory range filters
	kops toolbox instance-selector --instance-group-name ondemand-ig --vcpus-min=2 --vcpus-max=4 --memory-min 2048 --memory-max 4096
	`))

	toolboxInstanceSelectorShort = i18n.T(`Generate on-demand or spot instance-group specs by providing resource specs like vcpus and memory.`)
)

// NewCmdToolboxInstanceSelector defines the cobra command for the instance-selector tool
func NewCmdToolboxInstanceSelector(f *util.Factory, out io.Writer) *cobra.Command {
	commandline := cli.New(
		"instance-selector",
		toolboxInstanceSelectorShort,
		toolboxInstanceSelectorLong,
		toolboxInstanceSelectorExample,
		func(cmd *cobra.Command, args []string) {},
	)
	commandline.Command.Run = func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()

		if err := rootCommand.ProcessArgs(args); err != nil {
			exitWithError(err)
		}
		err := RunToolboxInstanceSelector(ctx, f, out, rootCommand.ClusterName(), &commandline)
		if err != nil {
			exitWithError(err)
		}
	}

	cpuArchs := []string{"x86_64", "i386", "arm64"}
	cpuArchDefault := "x86_64"
	placementGroupStrategies := []string{"cluster", "partition", "spread"}
	usageClasses := []string{usageClassSpot, usageClassOndemand}
	usageClassDefault := usageClassOndemand
	outputDefault := "yaml"
	dryRunDefault := false
	clusterAutoscalerDefault := false
	nodeCountMinDefault := 2
	nodeCountMaxDefault := 15
	maxResultsDefault := 20

	commandline.StringFlag(igName, nil, nil, "Name of the Instance-Group", func(val interface{}) error {
		if val == nil {
			return fmt.Errorf("error you must supply --%s", igName)
		}
		matched, err := regexp.MatchString(nameRegex, *val.(*string))
		if err != nil {
			return err
		}
		if matched {
			return nil
		}
		return fmt.Errorf("error --%s must conform to the regex: \"%s\"", igName, nameRegex)
	})

	// Instance Group Node Configurations

	commandline.IntFlag(nodeCountMin, nil, &nodeCountMinDefault, "Set the minimum number of nodes")
	commandline.IntFlag(nodeCountMax, nil, &nodeCountMaxDefault, "Set the maximum number of nodes")
	commandline.IntFlag(nodeVolumeSize, nil, nil, "Set instance volume size (in GiB) for nodes")
	commandline.StringSliceFlag(nodeSecurityGroups, nil, nil, "Add precreated additional security groups to nodes")
	commandline.BoolFlag(clusterAutoscaler, nil, &clusterAutoscalerDefault, "Add auto-discovery tags for cluster-autoscaler to manage the instance-group")

	// Aggregate Filters

	commandline.StringFlag(instanceTypeBase, nil, nil, "Base instance type to retrieve similarly spec'd instance types", nil)
	commandline.BoolFlag(flexible, nil, nil, "Retrieves a group of instance types spanning multiple generations based on opinionated defaults and user overridden resource filters")
	commandline.IntFlag(instanceGroupCount, nil, nil, "Number of instance groups to create w/ different vcpus-to-memory-ratios starting at 1:2 and doubling.")

	// Raw Filters

	commandline.IntMinMaxRangeFlags(vcpus, nil, nil, "Number of vcpus available to the instance type.")
	commandline.IntMinMaxRangeFlags(memory, nil, nil, "Amount of memory available in MiB (Example: 4096)")
	commandline.RatioFlag(vcpusToMemoryRatio, nil, nil, "The ratio of vcpus to memory in MiB. (Example: 1:2)")
	commandline.StringOptionsFlag(cpuArchitecture, nil, &cpuArchDefault, fmt.Sprintf("CPU architecture [%s]", strings.Join(cpuArchs, ", ")), cpuArchs)
	commandline.IntMinMaxRangeFlags(gpus, nil, nil, "Total number of GPUs (Example: 4)")
	commandline.IntMinMaxRangeFlags(gpuMemoryTotal, nil, nil, "Number of GPUs' total memory in MiB (Example: 4096)")
	commandline.StringOptionsFlag(placementGroupStrategy, nil, nil, fmt.Sprintf("Placement group strategy: [%s]", strings.Join(placementGroupStrategies, ", ")), placementGroupStrategies)
	commandline.StringOptionsFlag(usageClass, nil, &usageClassDefault, fmt.Sprintf("Usage class: [%s]", strings.Join(usageClasses, ", ")), usageClasses)
	commandline.BoolFlag(enaSupport, nil, nil, "Instance types where ENA is supported or required")
	commandline.BoolFlag(burstSupport, nil, nil, "Burstable instance types")
	commandline.StringSliceFlag(availabilityZones, nil, nil, "Availability zones or zone ids to check only EC2 capacity offered in those specific AZs")
	commandline.IntMinMaxRangeFlags(networkInterfaces, nil, nil, "Number of network interfaces (ENIs) that can be attached to the instance")
	commandline.RegexFlag(allowList, nil, nil, "List of allowed instance types to select from w/ regex syntax (Example: m[3-5]\\.*)")
	commandline.RegexFlag(denyList, nil, nil, "List of instance types which should be excluded w/ regex syntax (Example: m[1-2]\\.*)")

	// Output Flags

	commandline.IntFlag(maxResults, nil, &maxResultsDefault, "Maximum number of instance types to return back")
	commandline.BoolFlag(dryRun, nil, &dryRunDefault, "If true, only print the object that would be sent, without sending it. This flag can be used to create a cluster YAML or JSON manifest.")
	commandline.StringFlag(output, commandline.StringMe("o"), &outputDefault, "Output format. One of json|yaml. Used with the --dry-run flag.", nil)

	return commandline.Command
}

// RunToolboxInstanceSelector executes the instance-selector tool to create instance groups with declarative resource specifications
func RunToolboxInstanceSelector(ctx context.Context, f *util.Factory, out io.Writer, clusterName string, commandline *cli.CommandLineInterface) error {

	flags, err := processAndValidateFlags(commandline, clusterName)
	if err != nil {
		return err
	}
	instanceSelectorOpts := getInstanceSelectorOpts(commandline)

	clientset, cluster, channel, err := retrieveClusterRefs(ctx, f, clusterName)
	if err != nil {
		return err
	}

	zones, err := getClusterZones(cluster.Spec.Subnets)
	if err != nil {
		return err
	}
	region := zones[0][:len(zones[0])-1]

	tags := map[string]string{"KubernetesCluster": clusterName}
	cloud, err := awsup.NewAWSCloud(region, tags)
	if err != nil {
		return fmt.Errorf("error initializing AWS client: %v", err)
	}

	instanceSelector := selector.Selector{
		EC2: cloud.EC2(),
	}

	igCount := instanceSelectorOpts.InstanceGroupCount
	if flags[instanceGroupCount] == nil {
		igCount = 1
	}
	filters := getFilters(commandline, region)
	mutatedFilters := filters
	if flags[instanceGroupCount] != nil || filters.Flexible != nil {
		if filters.VCpusToMemoryRatio == nil {
			defaultStartRatio := float64(2.0)
			mutatedFilters.VCpusToMemoryRatio = &defaultStartRatio
		}
	}

	instanceGroupName := instanceSelectorOpts.InstanceGroupName
	newInstanceGroups := []*kops.InstanceGroup{}

	for i := 0; i < igCount; i++ {
		igNameForRun := instanceGroupName
		if igCount != 1 {
			igNameForRun = fmt.Sprintf("%s%d", instanceGroupName, i+1)
		}
		selectedInstanceTypes, err := instanceSelector.Filter(mutatedFilters)
		if err != nil {
			return fmt.Errorf("error finding matching instance types: %w", err)
		}
		if len(selectedInstanceTypes) == 0 {
			return fmt.Errorf("no instance types were returned because the criteria specified was too narrow")
		}
		usageClass := *filters.UsageClass

		ig := createInstanceGroup(igNameForRun, clusterName, zones)
		ig = decorateWithInstanceGroupSpecs(ig, instanceSelectorOpts)
		ig, err = decorateWithMixedInstancesPolicy(ig, usageClass, selectedInstanceTypes)
		if err != nil {
			return err
		}
		if instanceSelectorOpts.ClusterAutoscaler {
			ig = decorateWithClusterAutoscalerLabels(ig)
		}
		ig, err = cloudup.PopulateInstanceGroupSpec(cluster, ig, channel)
		if err != nil {
			return err
		}

		newInstanceGroups = append(newInstanceGroups, ig)

		if igCount != 1 {
			doubledRatio := (*mutatedFilters.VCpusToMemoryRatio) * 2
			mutatedFilters.VCpusToMemoryRatio = &doubledRatio
		}
	}

	if instanceSelectorOpts.DryRun {
		for _, ig := range newInstanceGroups {
			switch instanceSelectorOpts.Output {
			case OutputYaml:
				if err := fullOutputYAML(out, ig); err != nil {
					return fmt.Errorf("error writing cluster yaml to stdout: %v", err)
				}
			case OutputJSON:
				if err := fullOutputJSON(out, ig); err != nil {
					return fmt.Errorf("error writing cluster json to stdout: %v", err)
				}
			default:
				return fmt.Errorf("unsupported output type %q", instanceSelectorOpts.Output)
			}
		}
		return nil
	}

	for _, ig := range newInstanceGroups {
		_, err = clientset.InstanceGroupsFor(cluster).Create(ctx, ig, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("error storing InstanceGroup: %v", err)
		}

		if err := fullOutputYAML(out, ig); err != nil {
			return fmt.Errorf("error writing cluster yaml to stdout: %v", err)
		}
	}

	return nil
}

func processAndValidateFlags(commandline *cli.CommandLineInterface, clusterName string) (map[string]interface{}, error) {
	if err := commandline.SetUntouchedFlagValuesToNil(); err != nil {
		return nil, err
	}

	if err := commandline.ProcessRangeFilterFlags(); err != nil {
		return nil, err
	}

	if err := commandline.ValidateFlags(); err != nil {
		return nil, err
	}

	if clusterName == "" {
		return nil, fmt.Errorf("ClusterName is required")
	}

	return commandline.Flags, nil
}

func retrieveClusterRefs(ctx context.Context, f *util.Factory, clusterName string) (simple.Clientset, *kops.Cluster, *kops.Channel, error) {
	clientset, err := f.Clientset()
	if err != nil {
		return nil, nil, nil, err
	}

	cluster, err := clientset.GetCluster(ctx, clusterName)
	if err != nil {
		return nil, nil, nil, err
	}

	if cluster == nil {
		return nil, nil, nil, fmt.Errorf("cluster %q not found", clusterName)
	}

	channel, err := cloudup.ChannelForCluster(cluster)
	if err != nil {
		return nil, nil, nil, err
	}

	if len(cluster.Spec.Subnets) == 0 {
		return nil, nil, nil, fmt.Errorf("configuration must include Subnets")
	}

	return clientset, cluster, channel, nil
}

func getFilters(commandline *cli.CommandLineInterface, region string) selector.Filters {
	flags := commandline.Flags
	return selector.Filters{
		VCpusRange:             commandline.IntRangeMe(flags[vcpus]),
		MemoryRange:            commandline.IntRangeMe(flags[memory]),
		VCpusToMemoryRatio:     commandline.Float64Me(flags[vcpusToMemoryRatio]),
		CPUArchitecture:        commandline.StringMe(flags[cpuArchitecture]),
		GpusRange:              commandline.IntRangeMe(flags[gpus]),
		GpuMemoryRange:         commandline.IntRangeMe(flags[gpuMemoryTotal]),
		PlacementGroupStrategy: commandline.StringMe(flags[placementGroupStrategy]),
		UsageClass:             commandline.StringMe(flags[usageClass]),
		EnaSupport:             commandline.BoolMe(flags[enaSupport]),
		Burstable:              commandline.BoolMe(flags[burstSupport]),
		Region:                 commandline.StringMe(region),
		AvailabilityZones:      commandline.StringSliceMe(flags[availabilityZones]),
		MaxResults:             commandline.IntMe(flags[maxResults]),
		NetworkInterfaces:      commandline.IntRangeMe(flags[networkInterfaces]),
		NetworkPerformance:     commandline.IntRangeMe(flags[networkPerformance]),
		AllowList:              commandline.RegexMe(flags[allowList]),
		DenyList:               commandline.RegexMe(flags[denyList]),
		InstanceTypeBase:       commandline.StringMe(flags[instanceTypeBase]),
		Flexible:               commandline.BoolMe(flags[flexible]),
	}
}

func getInstanceSelectorOpts(commandline *cli.CommandLineInterface) InstanceSelectorOptions {
	opts := InstanceSelectorOptions{}
	flags := commandline.Flags
	opts.NodeCountMin = int32(*commandline.IntMe(flags[nodeCountMin]))
	opts.NodeCountMax = int32(*commandline.IntMe(flags[nodeCountMax]))
	opts.InstanceGroupName = *commandline.StringMe(flags[igName])
	opts.Output = *commandline.StringMe(flags[output])
	opts.DryRun = *commandline.BoolMe(flags[dryRun])
	opts.ClusterAutoscaler = *commandline.BoolMe(flags[clusterAutoscaler])
	if flags[nodeVolumeSize] != nil {
		volumeSize := int32(*commandline.IntMe(flags[nodeVolumeSize]))
		opts.NodeVolumeSize = &volumeSize
	}
	if flags[nodeSecurityGroups] != nil {
		opts.NodeSecurityGroups = *commandline.StringSliceMe(flags[nodeSecurityGroups])
	}
	if flags[instanceGroupCount] != nil {
		opts.InstanceGroupCount = *commandline.IntMe(flags[instanceGroupCount])
	}
	return opts
}

func getClusterZones(subnets []kops.ClusterSubnetSpec) ([]string, error) {
	region := ""
	zones := []string{}
	for _, subnet := range subnets {
		zoneRegion := subnet.Zone[:len(subnet.Zone)-1]
		zones = append(zones, subnet.Zone)
		if region != "" && zoneRegion != region {
			return nil, fmt.Errorf("clusters cannot span multiple regions")
		}
		region = zoneRegion
	}
	if len(zones) == 0 {
		return nil, fmt.Errorf("the cluster must include at least 1 subnet")
	}
	return zones, nil
}

func createInstanceGroup(groupName, clusterName string, zones []string) *kops.InstanceGroup {
	ig := &kops.InstanceGroup{}
	ig.ObjectMeta.Name = groupName
	ig.Spec.Role = kops.InstanceGroupRoleNode
	ig.Spec.Subnets = zones
	ig.ObjectMeta.Labels = make(map[string]string)
	ig.ObjectMeta.Labels[kops.LabelClusterName] = clusterName

	ig.AddInstanceGroupNodeLabel()
	return ig
}

func decorateWithInstanceGroupSpecs(instanceGroup *kops.InstanceGroup, instanceGroupOpts InstanceSelectorOptions) *kops.InstanceGroup {
	ig := instanceGroup
	ig.Spec.MinSize = &instanceGroupOpts.NodeCountMin
	ig.Spec.MaxSize = &instanceGroupOpts.NodeCountMax
	ig.Spec.RootVolumeSize = instanceGroupOpts.NodeVolumeSize
	ig.Spec.AdditionalSecurityGroups = instanceGroupOpts.NodeSecurityGroups
	return ig
}

func decorateWithMixedInstancesPolicy(instanceGroup *kops.InstanceGroup, usageClass string, instanceSelections []string) (*kops.InstanceGroup, error) {
	ig := instanceGroup
	ig.Spec.MachineType = instanceSelections[0]

	if usageClass == usageClassSpot {
		ondemandBase := int64(0)
		ondemandAboveBase := int64(0)
		spotAllocationStrategy := "capacity-optimized"
		ig.Spec.MixedInstancesPolicy = &kops.MixedInstancesPolicySpec{
			Instances:              instanceSelections,
			OnDemandBase:           &ondemandBase,
			OnDemandAboveBase:      &ondemandAboveBase,
			SpotAllocationStrategy: &spotAllocationStrategy,
		}
	} else if usageClass == usageClassOndemand {
		ig.Spec.MixedInstancesPolicy = &kops.MixedInstancesPolicySpec{
			Instances: instanceSelections,
		}
	} else {
		return nil, fmt.Errorf("error node usage class not supported")
	}

	generatedWithLabelKey := "kops.k8s.io/instance-selector"
	if ig.Spec.CloudLabels == nil {
		ig.Spec.CloudLabels = make(map[string]string)
	}
	ig.Spec.CloudLabels[generatedWithLabelKey] = "1"

	return ig, nil
}

func decorateWithClusterAutoscalerLabels(instanceGroup *kops.InstanceGroup) *kops.InstanceGroup {
	ig := instanceGroup
	clusterName := instanceGroup.ObjectMeta.Name
	if ig.Spec.CloudLabels == nil {
		ig.Spec.CloudLabels = make(map[string]string)
	}
	ig.Spec.CloudLabels["k8s.io/cluster-autoscaler/enabled"] = igName
	ig.Spec.CloudLabels["k8s.io/cluster-autoscaler/"+clusterName] = igName
	return ig
}
