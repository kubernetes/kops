package main

type kubectlConfig struct {
	Kind       string                    `json:"kind`
	ApiVersion string                    `json:"apiVersion`
	Clusters   []*kubectlClusterWithName `json:"clusters`
}

type kubectlClusterWithName struct {
	Name    string         `json:"name`
	Cluster kubectlCluster `json:"cluster`
}

type kubectlCluster struct {
	Server string `json:"server`
}
