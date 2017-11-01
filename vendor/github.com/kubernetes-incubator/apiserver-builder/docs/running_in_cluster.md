# Running an apiserver and controller-manager in-cluster

This document describes how to run an apiserver and controller-manager
in cluster using API aggregation.

For details on the underlying concepts applied by the commands,
see the [auth concept page](concepts/auth.md).

## TL;DR version

Running the following command will automatically invoke each of the commands
covered in the [long version](#long-version)

`apiserver-boot build config --name <servicename> --namespace <namespace to run in> --image <image to run>`

## Long version

### Build the container image

**Note:** If your apiserver and controller-manager were not created
with the apiserver-builder framework, you made need to manually build
your container image.

`apiserver-boot build container --image <image>`

This will generate code, build the apiserver and controller-manager
binaries and then build a container image.

Push the image with:

`docker push <image>`

### Build the config

`apiserver-boot build config --name <servicename> --namespace <namespace to run in> --image <image to run>`

This will perform the following:

- create a CA + certificate for the service to use
  - under config/certificates
- locate each API group/version based on the directory structure
- create config for the APIServices, Deployment, Service, and Secret
  - in config/apiserver.yaml

**Note:** This relies on the container have the binaries `apiserver` and `controller-manager`
present and runnable from "./".  You may need to manually edit the config if your
container looks differently.

### Run the apiserver

`kubectl apply -f config/apiserver.yaml`

Clear the discovery cache:

`rm -rf ~/.kube/cache/discovery/`

Look for your API:

`kubectl api-versions`

## Create an instance of your resource

`kubectl apply -f sample/<type>.yaml`