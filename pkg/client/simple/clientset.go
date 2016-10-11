package simple

type Clientset interface {
	Clusters() ClusterInterface
	InstanceGroups(cluster string) InstanceGroupInterface
	Federations() FederationInterface
}