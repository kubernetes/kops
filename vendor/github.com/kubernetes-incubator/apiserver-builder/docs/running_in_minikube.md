# Running an apiserver and controller-manager in-cluster

This document describes how to run an apiserver and controller-manager
in locally but using API aggregation with a local minikube cluster.

## Before you start 

Make sure your repo is setup

- Install minikube and start a new cluster
- Create a new GO project
- In the GO project init the repo with `apiserver-boot init`
- Create a resource with `apiserver-boot create group version resource`

## Start minikube

Make sure you are using a cluster at least 1.7.5

`minikube start`

## Build the aggregation config for your minikube cluster

`apiserver-boot build config --local-minikube --name <servicename> --namespace <namespace to run in>`

## Configure the minikube cluster

`kubectl create -f config/apiserver.yaml`

## Run you apiserver and controller-manager locally aggregated with the minikube cluster

`apiserver-boot run local-minikube`

If you have [Bazel](https://bazel.build/) & [Gazelle](https://github.com/bazelbuild/rules_go/tree/master/go/tools/gazelle)
installed, you can use the `--bazel` and `--gazelle` to build your binaries.  (Much faster!)
