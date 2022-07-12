#!/usr/bin/env bash

KOPS_VERSION=`.build/dist/$(go env GOOS)/$(go env GOARCH)/kops version -- --short`
export DOCKER_IMAGE_PREFIX=kops/
export DOCKER_REGISTRY=rg.fr-par.scw.cloud

if [[ $1 == "dns" ]] || [[ $2 == "dns" ]]
then
  make dns-controller-push
  export DNSCONTROLLER_IMAGE=${DOCKER_IMAGE_PREFIX}dns-controller:${KOPS_VERSION}
fi

if [[ $1 == "kops" ]] || [[ $2 == "kops" ]]
then
  make kops-controller-push
  export KOPSCONTROLLER_IMAGE=${DOCKER_IMAGE_PREFIX}kops-controller:${KOPS_VERSION}
fi