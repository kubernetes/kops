## `apiserver-builder`

[![Build Status](https://travis-ci.org/kubernetes-incubator/apiserver-builder.svg?branch=master)](https://travis-ci.org/kubernetes-incubator/apiserver-builder "Travis")
[![Go Report Card](https://goreportcard.com/badge/github.com/kubernetes-incubator/apiserver-builder)](https://goreportcard.com/report/github.com/kubernetes-incubator/apiserver-builder)

**Note**: This project is still only a proof of concept, and is not production ready.

Apiserver Builder is a collection of libraries and tools to build native
Kubernetes extensions using Kubernetes apiserver code.

## Motivation

*Addon apiservers* are a Kubernetes extension point allowing fully featured Kubernetes
APIs to be developed on the same api-machinery used to build the core Kubernetes APIS,
but with the flexibility of being distributed and installed separately from
the Kubernetes project.  This allows APIs to be developed outside of the
Kubernetes repo and installed separately as a package.

Building addon apiservers directly on the raw api-machinery libraries requires non-trivial
code that must be maintained and rebased as the raw libraries change. The goal of this project is
to make building apiservers in *Go* simple and accessible to everyone in the
Kubernetes community.

apiserver-builder provides libraries, code generators, and tooling to make it possible to build
and run a basic apiserver in an afternoon, while providing all of the hooks to offer the
same capabilities when building from scratch.

## Highlights

- Tools to bootstrap type definitions, controllers, tests and documentation for new resources
- Tools to build and run the extension control plane standalone and in minikube and remote clusters.
- Easily watch and update Kubernetes API types from your controller
- Easily add new resources and subresources
- Provides sane defaults for most values, but can be overridden

## Guides

**Note:** The guides are presented roughly in the order of recommended progression.

#### Installation guide

Download the latest release and install on your PATH.

[installation guide](docs/installing.md)

#### Building APIs concept guide

Conceptual information on how APIs and the Kubernetes control plane is structure and how to
build new API extensions using apiserver-builder.

If you want to get straight to building something without knowing all the details of what is going on,
skip ahead to the tools guide and come back to this later.

[api building concept guide](docs/concepts/api_building_overview.md)


#### Tools user guide

Instructions on how to use the tools packaged with apiserver-builder to build and run a new apiserver.

[tools guide](docs/tools_user_guide.md)

#### Step by step example

List of commits showing `apiserver-boot` commands run and the corresponding changes:

https://github.com/kubernetes-incubator/apiserver-builder/commits/example-simple

#### Coding and libraries user guide

Instructions for how to implement custom APIs on top of the apiserver-builder libraries.

[libraries guide](docs/libraries_user_guide.md)

#### Concept guides

Conceptual information on addon apiservers, such as how auth works and how they interact
with the main Kubernetes API server and API aggregator.

[Concepts](docs/concepts/README.md)

## Additional material

##### Using delegated auth with minikube

Instructions on how to run an apiserver using delegated auth with a minikube cluster

Details [here](https://github.com/kubernetes-incubator/apiserver-builder/blob/master/docs/using_minikube.md)
