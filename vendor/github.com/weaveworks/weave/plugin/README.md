# Weave network driver extension for Docker and Kubernetes

This directory implements both a 
[Docker plugin](http://docs.docker.com/engine/extend/plugin_api/) to
integrate [Weave Net](http://weave.works/net/) with Docker, and a 
[Container Network Interface (CNI) plugin](https://github.com/appc/cni#cni---the-container-network-interface),
to integrate Weave Net with [Kubernetes](http://kubernetes.io/).

The Docker plugin runs automatically when you `weave launch`, provided your
Docker daemon is version 1.9 or newer.

More detail on the Docker plugin [here](https://www.weave.works/docs/net/latest/plugin/).

More detail on the CNI plugin [here](https://www.weave.works/docs/net/latest/cni-plugin/).
