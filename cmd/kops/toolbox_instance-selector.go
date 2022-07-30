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
	"strings"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/cli"
	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/selector"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/commands/commandutils"
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
	gpuMemory              = "gpu-memory"
	placementGroupStrategy = "placement-group-strategy"
	usageClass             = "usage-class"
	enaSupport             = "ena-support"
	burstSupport           = "burst-support"
	subnets                = "subnets"
	networkInterfaces      = "network-interfaces"
	networkPerformance     = "network-performance"
	allowList              = "allow-list"
	denyList               = "deny-list"
	maxResults             = "max-results"
)

// Aggregate Filter Flag Constants
const (
	instanceTypeBase = "base-instance-type"
	flexible         = "flexible"
)

// Control Flag Constants
const (
	instanceGroupCount   = "ig-count"
	nodeCountMin         = "node-count-min"
	nodeCountMax         = "node-count-max"
	nodeVolumeSize       = "node-volume-size"
	nodeSecurityGroups   = "node-security-groups"
	clusterAutoscaler    = "cluster-autoscaler"
	usageClassSpot       = "spot"
	usageClassOndemand   = "on-demand"
	dryRun               = "dry-run"
	output               = "output"
	cpuArchitectureAMD64 = "amd64"
	cpuArchitectureX8664 = "x86_64"
	cpuArchitectureARM64 = "arm64"
)

// InstanceSelectorOptions is a struct representing non-filter flags passed into instance-selector
type InstanceSelectorOptions struct {
	ClusterName        string
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
	Generate AWS EC2 instance groups by providing resource specs,
	such as vcpus and memory, rather than instance types.`))

	toolboxInstanceSelectorExample = templates.Examples(i18n.T(`
	## Create a spot instance group using a MixInstancesPolicy and Capacity-Optimized spot allocation strategy.
	## --flexible defaults to a 1:2 vcpus to memory ratio and 4 vcpus.
	kops toolbox instance-selector my-spot-mig --usage-class spot --flexible

	## Create an on-demand instance group with custom vcpu and memory range filters.
	kops toolbox instance-selector ondemand-ig --vcpus-min=2 --vcpus-max=4 --memory-min 2gb --memory-max 4gb
	`))

	toolboxInstanceSelectorShort = i18n.T(`Generate instance-group specs by providing resource specs such as vcpus and memory.`)
)

// NewCmdToolboxInstanceSelector defines the cobra command for the instance-selector tool
func NewCmdToolboxInstanceSelector(f commandutils.Factory, out io.Writer) *cobra.Command {
	options := &InstanceSelectorOptions{}

	commandline := cli.New(
		"instance-selector INSTANCE_GROUP",
		toolboxInstanceSelectorShort,
		toolboxInstanceSelectorLong,
		toolboxInstanceSelectorExample,
		nil,
	)
	commandline.Command.Args = func(cmd *cobra.Command, args []string) error {
		options.ClusterName = rootCommand.ClusterName(true)

		if options.ClusterName == "" {
			return fmt.Errorf("--name is required")
		}

		if len(args) == 0 {
			return fmt.Errorf("must specify name of instance group to create")
		}

		options.InstanceGroupName = args[0]

		if len(args) != 1 {
			return fmt.Errorf("can only create one instance group at a time")
		}

		if err := processAndValidateFlags(&commandline); err != nil {
			return err
		}
		setInstanceSelectorOpts(options, &commandline)

		return nil
	}
	commandline.Command.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		commandutils.ConfigureKlogForCompletion()
		if len(args) == 1 && rootCommand.ClusterName(false) == "" {
			return []string{"--name"}, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	commandline.Command.RunE = func(cmd *cobra.Command, args []string) error {
		return RunToolboxInstanceSelector(context.TODO(), f, out, &commandline, options)
	}

	cpuArchs := []string{cpuArchitectureAMD64, cpuArchitectureARM64}
	cpuArchDefault := cpuArchitectureAMD64
	placementGroupStrategies := []string{"cluster", "partition", "spread"}
	usageClasses := []string{usageClassSpot, usageClassOndemand}
	usageClassDefault := usageClassOndemand
	outputDefault := "yaml"
	dryRunDefault := false
	clusterAutoscalerDefault := true
	nodeCountMinDefault := 1
	nodeCountMaxDefault := 10
	maxResultsDefault := 20

	// Instance Group Node Configurations

	commandline.IntFlag(nodeCountMin, nil, &nodeCountMinDefault, "Minimum number of nodes")
	commandline.IntFlag(nodeCountMax, nil, &nodeCountMaxDefault, "Maximum number of nodes")
	commandline.IntFlag(nodeVolumeSize, nil, nil, "Instance volume size (in GiB) for nodes")
	commandline.StringSliceFlag(nodeSecurityGroups, nil, nil, "Pre-created additional security groups for nodes")
	commandline.Command.RegisterFlagCompletionFunc(nodeSecurityGroups, completeSecurityGroup)
	commandline.BoolFlag(clusterAutoscaler, nil, &clusterAutoscalerDefault, "Add auto-discovery tags for cluster-autoscaler to manage the instance-group")

	// Aggregate Filters

	commandline.StringFlag(instanceTypeBase, nil, nil, "Base instance type to retrieve similarly specified instance types", nil)
	commandline.Command.RegisterFlagCompletionFunc(instanceTypeBase, completeMachineType)
	commandline.BoolFlag(flexible, nil, nil, "Retrieve a group of instance types spanning multiple generations based on opinionated defaults and user overridden resource filters")
	commandline.IntFlag(instanceGroupCount, nil, nil, "Number of instance groups to create with different vcpus-to-memory ratios, starting at 1:2 and doubling")

	// Raw Filters

	commandline.IntMinMaxRangeFlags(vcpus, nil, nil, "Number of vcpus available to the instance type")
	commandline.ByteQuantityMinMaxRangeFlags(memory, nil, nil, "Amount of memory available (Example: 4gb)")
	commandline.Command.RegisterFlagCompletionFunc(memory, cobra.NoFileCompletions)
	commandline.Command.RegisterFlagCompletionFunc(memory+"-min", cobra.NoFileCompletions)
	commandline.Command.RegisterFlagCompletionFunc(memory+"-max", cobra.NoFileCompletions)
	commandline.RatioFlag(vcpusToMemoryRatio, nil, nil, "Ratio of vcpus to memory in MiB. (Example: 1:2)")
	commandline.Command.RegisterFlagCompletionFunc(vcpusToMemoryRatio, cobra.NoFileCompletions)
	commandline.StringOptionsFlag(cpuArchitecture, nil, &cpuArchDefault, fmt.Sprintf("CPU architecture [%s]", strings.Join(cpuArchs, ", ")), append(cpuArchs, cpuArchitectureX8664))
	commandline.Command.RegisterFlagCompletionFunc(cpuArchitecture, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return cpuArchs, cobra.ShellCompDirectiveNoFileComp
	})
	commandline.IntMinMaxRangeFlags(gpus, nil, nil, "Number of GPUs (Example: 4)")
	commandline.ByteQuantityMinMaxRangeFlags(gpuMemory, nil, nil, "GPUs' total memory (Example: 4gb)")
	commandline.Command.RegisterFlagCompletionFunc(gpuMemory, cobra.NoFileCompletions)
	commandline.Command.RegisterFlagCompletionFunc(gpuMemory+"-min", cobra.NoFileCompletions)
	commandline.Command.RegisterFlagCompletionFunc(gpuMemory+"-max", cobra.NoFileCompletions)
	commandline.StringOptionsFlag(placementGroupStrategy, nil, nil, fmt.Sprintf("Placement group strategy: [%s]", strings.Join(placementGroupStrategies, ", ")), placementGroupStrategies)
	commandline.Command.RegisterFlagCompletionFunc(placementGroupStrategy, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return placementGroupStrategies, cobra.ShellCompDirectiveNoFileComp
	})
	commandline.StringOptionsFlag(usageClass, nil, &usageClassDefault, fmt.Sprintf("Usage class: [%s]", strings.Join(usageClasses, ", ")), usageClasses)
	commandline.Command.RegisterFlagCompletionFunc(usageClass, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return usageClasses, cobra.ShellCompDirectiveNoFileComp
	})
	commandline.BoolFlag(enaSupport, nil, nil, "Instance types where ENA is supported or required")
	commandline.BoolFlag(burstSupport, nil, nil, "Burstable instance types")
	commandline.StringSliceFlag(subnets, nil, nil, "Subnet(s) in which to create the instance group. One of Availability Zone like eu-west-1a or utility-eu-west-1a,")
	commandline.Command.RegisterFlagCompletionFunc(subnets, completeClusterSubnet(f, commandline.Flags[subnets].(*[]string)))
	commandline.IntMinMaxRangeFlags(networkInterfaces, nil, nil, "Number of network interfaces (ENIs) that can be attached to the instance")
	commandline.RegexFlag(allowList, nil, nil, "List of allowed instance types to select from w/ regex syntax (Example: m[3-5]\\.*)")
	commandline.Command.RegisterFlagCompletionFunc(allowList, cobra.NoFileCompletions)
	commandline.RegexFlag(denyList, nil, nil, "List of instance types which should be excluded w/ regex syntax (Example: m[1-2]\\.*)")
	commandline.Command.RegisterFlagCompletionFunc(denyList, cobra.NoFileCompletions)

	// Output Flags

	commandline.IntFlag(maxResults, nil, &maxResultsDefault, "Maximum number of instance types to return back")
	commandline.BoolFlag(dryRun, nil, &dryRunDefault, "Only print the object that would be created, without creating it. This flag can be used to create a cluster YAML or JSON manifest.")
	commandline.StringFlag(output, commandline.StringMe("o"), &outputDefault, "Output format. One of json or yaml. Used with the --dry-run flag.", nil)
	commandline.Command.RegisterFlagCompletionFunc(output, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "yaml"}, cobra.ShellCompDirectiveNoFileComp
	})

	return commandline.Command
}

// RunToolboxInstanceSelector executes the instance-selector tool to create instance groups with declarative resource specifications
func RunToolboxInstanceSelector(ctx context.Context, f commandutils.Factory, out io.Writer, commandline *cli.CommandLineInterface, options *InstanceSelectorOptions) error {
	clientset, cluster, channel, err := retrieveClusterRefs(ctx, f, options.ClusterName)
	if err != nil {
		return err
	}

	if cluster.Spec.GetCloudProvider() != kops.CloudProviderAWS {
		return fmt.Errorf("cannot select instance types from non-aws cluster")
	}

	firstClusterSubnet := strings.ReplaceAll(cluster.Spec.Subnets[0].Name, "utility-", "")
	region := firstClusterSubnet[:len(firstClusterSubnet)-1]

	igSubnets := []string{}
	for _, clusterSubnet := range cluster.Spec.Subnets {
		igSubnets = append(igSubnets, clusterSubnet.Name)
	}

	if commandline.Flags[subnets] != nil {
		userSubnets := *commandline.StringSliceMe(commandline.Flags[subnets])
		dryRun := *commandline.BoolMe(commandline.Flags[dryRun])
		err := validateUserSubnets(userSubnets, cluster.Spec.Subnets)
		if err != nil && !dryRun {
			return err
		}
		igSubnets = userSubnets
	}

	zones := []string{}
	for _, igSubnet := range igSubnets {
		zones = append(zones, strings.ReplaceAll(igSubnet, "utility-", ""))
	}

	tags := map[string]string{"KubernetesCluster": options.ClusterName}
	cloud, err := awsup.NewAWSCloud(region, tags)
	if err != nil {
		return fmt.Errorf("error initializing AWS client: %v", err)
	}

	instanceSelector := selector.Selector{
		EC2: cloud.EC2(),
	}

	igCount := options.InstanceGroupCount
	filters := getFilters(commandline, region, zones)
	mutatedFilters := filters
	if commandline.Flags[instanceGroupCount] != nil || filters.Flexible != nil {
		if filters.VCpusToMemoryRatio == nil {
			defaultStartRatio := float64(2.0)
			mutatedFilters.VCpusToMemoryRatio = &defaultStartRatio
		}
	}

	newInstanceGroups := []*kops.InstanceGroup{}

	for i := 0; i < igCount; i++ {
		igNameForRun := options.InstanceGroupName
		if igCount != 1 {
			igNameForRun = fmt.Sprintf("%s%d", options.InstanceGroupName, i+1)
		}
		selectedInstanceTypes, err := instanceSelector.Filter(mutatedFilters)
		if err != nil {
			return fmt.Errorf("error finding matching instance types: %w", err)
		}
		if len(selectedInstanceTypes) == 0 {
			return fmt.Errorf("no instance types were returned because the criteria specified was too narrow")
		}
		usageClass := *filters.UsageClass

		ig := createInstanceGroup(igNameForRun, options.ClusterName, igSubnets)
		ig = decorateWithInstanceGroupSpecs(ig, options)
		ig, err = decorateWithMixedInstancesPolicy(ig, usageClass, selectedInstanceTypes)
		if err != nil {
			return err
		}
		if options.ClusterAutoscaler {
			ig = decorateWithClusterAutoscalerLabels(ig, options.ClusterName)
		}
		ig, err = cloudup.PopulateInstanceGroupSpec(cluster, ig, cloud, channel)
		if err != nil {
			return err
		}

		newInstanceGroups = append(newInstanceGroups, ig)

		if igCount != 1 {
			doubledRatio := (*mutatedFilters.VCpusToMemoryRatio) * 2
			mutatedFilters.VCpusToMemoryRatio = &doubledRatio
		}
	}

	if options.DryRun {
		for _, ig := range newInstanceGroups {
			switch options.Output {
			case OutputYaml:
				if err := fullOutputYAML(out, ig); err != nil {
					return fmt.Errorf("error writing cluster yaml to stdout: %v", err)
				}
			case OutputJSON:
				if err := fullOutputJSON(out, true, ig); err != nil {
					return fmt.Errorf("error writing cluster json to stdout: %v", err)
				}
			default:
				return fmt.Errorf("unsupported output type %q", options.Output)
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

func processAndValidateFlags(commandline *cli.CommandLineInterface) error {
	if err := commandline.SetUntouchedFlagValuesToNil(); err != nil {
		return err
	}

	if err := commandline.ProcessFlags(); err != nil {
		return err
	}

	if err := commandline.ValidateFlags(); err != nil {
		return err
	}

	return nil
}

func retrieveClusterRefs(ctx context.Context, f commandutils.Factory, clusterName string) (simple.Clientset, *kops.Cluster, *kops.Channel, error) {
	clientset, err := f.KopsClient()
	if err != nil {
		return nil, nil, nil, err
	}

	cluster, err := GetCluster(ctx, f, clusterName)
	if err != nil {
		return nil, nil, nil, err
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

func getFilters(commandline *cli.CommandLineInterface, region string, zones []string) selector.Filters {
	flags := commandline.Flags
	return selector.Filters{
		VCpusRange:             commandline.IntRangeMe(flags[vcpus]),
		MemoryRange:            commandline.ByteQuantityRangeMe(flags[memory]),
		VCpusToMemoryRatio:     commandline.Float64Me(flags[vcpusToMemoryRatio]),
		CPUArchitecture:        commandline.StringMe(flags[cpuArchitecture]),
		GpusRange:              commandline.IntRangeMe(flags[gpus]),
		GpuMemoryRange:         commandline.ByteQuantityRangeMe(flags[gpuMemory]),
		PlacementGroupStrategy: commandline.StringMe(flags[placementGroupStrategy]),
		UsageClass:             commandline.StringMe(flags[usageClass]),
		EnaSupport:             commandline.BoolMe(flags[enaSupport]),
		Burstable:              commandline.BoolMe(flags[burstSupport]),
		Region:                 commandline.StringMe(region),
		AvailabilityZones:      commandline.StringSliceMe(zones),
		MaxResults:             commandline.IntMe(flags[maxResults]),
		NetworkInterfaces:      commandline.IntRangeMe(flags[networkInterfaces]),
		NetworkPerformance:     commandline.IntRangeMe(flags[networkPerformance]),
		AllowList:              commandline.RegexMe(flags[allowList]),
		DenyList:               commandline.RegexMe(flags[denyList]),
		InstanceTypeBase:       commandline.StringMe(flags[instanceTypeBase]),
		Flexible:               commandline.BoolMe(flags[flexible]),
	}
}

func setInstanceSelectorOpts(options *InstanceSelectorOptions, commandline *cli.CommandLineInterface) {
	flags := commandline.Flags
	options.NodeCountMin = int32(*commandline.IntMe(flags[nodeCountMin]))
	options.NodeCountMax = int32(*commandline.IntMe(flags[nodeCountMax]))
	options.Output = *commandline.StringMe(flags[output])
	options.DryRun = *commandline.BoolMe(flags[dryRun])
	options.ClusterAutoscaler = *commandline.BoolMe(flags[clusterAutoscaler])
	if flags[nodeVolumeSize] != nil {
		volumeSize := int32(*commandline.IntMe(flags[nodeVolumeSize]))
		options.NodeVolumeSize = &volumeSize
	}
	if flags[nodeSecurityGroups] != nil {
		options.NodeSecurityGroups = *commandline.StringSliceMe(flags[nodeSecurityGroups])
	}
	options.InstanceGroupCount = 1
	if flags[instanceGroupCount] != nil {
		options.InstanceGroupCount = *commandline.IntMe(flags[instanceGroupCount])
	}
}

func validateUserSubnets(userSubnets []string, clusterSubnets []kops.ClusterSubnetSpec) error {
	err := validateUserSubnetsWithClusterSubnets(userSubnets, clusterSubnets)
	if err != nil {
		return err
	}
	err = validateAllPrivateOrPublicSubnets(userSubnets)
	if err != nil {
		return err
	}
	return nil
}

// validateUserSubnetsWithClusterSubnets makes sure the userSubnets are part of the cluster subnets
func validateUserSubnetsWithClusterSubnets(userSubnets []string, clusterSubnets []kops.ClusterSubnetSpec) error {
	for _, userSubnet := range userSubnets {
		userSubnetValid := false
		for _, clusterSubnet := range clusterSubnets {
			if clusterSubnet.Name == userSubnet {
				userSubnetValid = true
				break
			}
		}
		if !userSubnetValid {
			return fmt.Errorf("error subnets must exist in the cluster")
		}
	}
	return nil
}

// validateAllPrivateOrPublicSubnets makes sure the passed in subnets are all utility (public) subnets or private subnets
func validateAllPrivateOrPublicSubnets(userSubnets []string) error {
	utilitySubnets := 0
	for _, userSubnet := range userSubnets {
		if strings.HasPrefix(userSubnet, "utility-") {
			utilitySubnets++
		}
	}

	if utilitySubnets != 0 && len(userSubnets) != utilitySubnets {
		return fmt.Errorf("error instance group cannot span public and private subnets")
	}
	return nil
}

func createInstanceGroup(groupName, clusterName string, subnets []string) *kops.InstanceGroup {
	ig := &kops.InstanceGroup{}
	ig.ObjectMeta.Name = groupName
	ig.Spec.Role = kops.InstanceGroupRoleNode
	ig.Spec.Subnets = subnets
	ig.ObjectMeta.Labels = make(map[string]string)
	ig.ObjectMeta.Labels[kops.LabelClusterName] = clusterName

	ig.AddInstanceGroupNodeLabel()
	return ig
}

func decorateWithInstanceGroupSpecs(instanceGroup *kops.InstanceGroup, instanceGroupOpts *InstanceSelectorOptions) *kops.InstanceGroup {
	ig := instanceGroup
	ig.Spec.MinSize = &instanceGroupOpts.NodeCountMin
	ig.Spec.MaxSize = &instanceGroupOpts.NodeCountMax
	ig.Spec.RootVolumeSize = instanceGroupOpts.NodeVolumeSize
	ig.Spec.AdditionalSecurityGroups = instanceGroupOpts.NodeSecurityGroups
	return ig
}

// decorateWithMixedInstancesPolicy adds a mixed instance policy based on usageClass to the instance-group
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

	return ig, nil
}

// decorateWithClusterAutoscalerLabels adds cluster-autoscaler discovery tags to the cloudlabels slice
func decorateWithClusterAutoscalerLabels(instanceGroup *kops.InstanceGroup, clusterName string) *kops.InstanceGroup {
	ig := instanceGroup
	if ig.Spec.CloudLabels == nil {
		ig.Spec.CloudLabels = make(map[string]string)
	}
	ig.Spec.CloudLabels["k8s.io/cluster-autoscaler/enabled"] = "1"
	ig.Spec.CloudLabels["k8s.io/cluster-autoscaler/"+clusterName] = "1"
	return ig
}
