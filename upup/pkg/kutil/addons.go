package kutil

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"k8s.io/kube-deploy/upup/pkg/fi/vfs"
	"strings"
)

type ClusterAddons struct {
	APIEndpoint string
	SSHConfig   ssh.ClientConfig
}

type ClusterAddon struct {
	Name string
	Path string
}

func (c *ClusterAddons) AddonsPath() (vfs.Path, error) {
	// TODO: Close NodeSSH

	// TODO: What if endpoint is a load balancer?  Query cloud and try to find actual hosts?

	// TODO: What if multiple masters?

	hostname := c.APIEndpoint
	hostname = strings.TrimPrefix(hostname, "http://")
	hostname = strings.TrimPrefix(hostname, "https://")
	master := &NodeSSH{
		Hostname: hostname,
	}

	master.SSHConfig = c.SSHConfig

	root, err := master.Root()
	if err != nil {
		return nil, err
	}

	manifests := root.Join("etc", "kubernetes", "addons")
	return manifests, nil
}

func (c *ClusterAddons) ListAddons() (map[string]*ClusterAddon, error) {
	addonsPath, err := c.AddonsPath()
	if err != nil {
		return nil, err
	}
	files, err := addonsPath.ReadDir()
	if err != nil {
		return nil, fmt.Errorf("error reading addons: %v", err)
	}

	addons := make(map[string]*ClusterAddon)

	for _, f := range files {
		name := f.Base()

		addon := &ClusterAddon{
			Name: name,
			Path: name,
		}
		addons[addon.Name] = addon
	}
	return addons, nil
}

func (c *ClusterAddons) CreateAddon(key string, files []vfs.Path) error {
	addonsPath, err := c.AddonsPath()
	if err != nil {
		return err
	}

	addonPath := addonsPath.Join(key)
	existingFiles, err := addonPath.ReadDir()
	if err == nil && len(existingFiles) != 0 {
		return fmt.Errorf("addon %q already exists", key)
	}

	srcData := make(map[string][]byte)
	for _, f := range files {
		name := f.Base()

		data, err := f.ReadFile()
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", f, err)
		}
		srcData[name] = data
	}

	for k, data := range srcData {
		destPath := addonPath.Join(k)
		err := destPath.WriteFile(data)
		if err != nil {
			// TODO: Delete other files?
			return fmt.Errorf("error writing file %s: %v", destPath, err)
		}
	}

	return nil
}
