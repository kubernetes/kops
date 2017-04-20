package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/api/types/swarm"
	"github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
)

const DOCKER_API_VERSION = "1.26"

func isSwarmManager(args []string) error {
	info, err := dockerInfo()
	if err != nil {
		return err
	}

	// hack-y way to denote that a node does not belong to any swarm
	if info.Swarm.LocalNodeState != swarm.LocalNodeStateActive {
		os.Exit(20)
	}

	if !info.Swarm.ControlAvailable {
		return fmt.Errorf("node is not a manager")
	}

	return nil
}

func swarmManagerPeers(args []string) error {
	info, err := dockerInfo()
	if err != nil {
		return err
	}

	for _, managerNode := range info.Swarm.RemoteManagers {
		ip, err := ipFromAddr(managerNode.Addr)
		if err != nil {
			return errors.Wrap(err, "ipFromAddr")
		}
		fmt.Println(ip)
	}

	return nil
}

func dockerInfo() (*docker.DockerInfo, error) {
	client, err := docker.NewVersionedClientFromEnv(DOCKER_API_VERSION)
	if err != nil {
		return nil, errors.Wrap(err, "docker.NewVersionedClientFromEnv")
	}

	info, err := client.Info()
	if err != nil {
		return nil, errors.Wrap(err, "docker.Info")
	}

	return info, nil
}

func ipFromAddr(addr string) (string, error) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid address: %q", addr)
	}

	return parts[0], nil
}
