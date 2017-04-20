package main

import (
	"fmt"
	"log"

	"k8s.io/client-go/kubernetes"
	api "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

func getKubePeers() ([]string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	nodeList, err := c.Nodes().List(api.ListOptions{})
	if err != nil {
		// Fallback for cases (e.g. from kube-up.sh) where kube-proxy is not running on master
		config.Host = "http://localhost:8080"
		log.Print("error contacting APIServer: ", err, "; trying with fallback: ", config.Host)
		c, err = kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}
		nodeList, err = c.Nodes().List(api.ListOptions{})
	}

	if err != nil {
		return nil, err
	}
	addresses := make([]string, 0, len(nodeList.Items))
	for _, peer := range nodeList.Items {
		for _, addr := range peer.Status.Addresses {
			if addr.Type == "InternalIP" {
				addresses = append(addresses, addr.Address)
			}
		}
	}
	return addresses, nil
}

func main() {
	peers, err := getKubePeers()
	if err != nil {
		log.Fatalf("Could not get peers: %v", err)
	}
	for _, addr := range peers {
		fmt.Println(addr)
	}
}
