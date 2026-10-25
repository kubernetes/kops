# Kindnet

***Kindnet support is experimental, and may be removed at any time***

## Introduction

* [kindnet](http://kindnet.es)

Kindnet focuses on providing essential networking functionality without unnecessary complexity.

## Installing

To install [kindnet](https://github.com/kubernetes-sigs/kindnet) - use `--networking kindnet`.

```sh
export ZONES=mylistofzone
kops create cluster \
  --zones $ZONES \
  --networking kindnet \
  --yes \
  --name myclustername.mydns.io
```

## Using a custom image
{{ kops_feature_table(kops_added_default='1.38') }}

kOps runs `registry.k8s.io/networking/kindnet:<version>`, where `<version>` comes from `spec.networking.kindnet.version`.
To run a different kindnet image, set `spec.networking.kindnet.image` to the full image reference. When `image` is set, kOps does not use `version`.

```yaml
spec:
  networking:
    kindnet:
      image: registry.example.com/kindnet:v1.0.2-custom
```

The same field can be set when creating the cluster:

```sh
kops create cluster \
  --networking kindnet \
  --set=cluster.spec.networking.kindnet.image=registry.example.com/kindnet:v1.0.2-custom \
  ...
```

## Getting help

For problems with kindnet please post an issue to Github:

- [Kindnet Issues](https://github.com/kubernetes-sigs/kindnet/issues)

You can learn more about the different configurations options in https://kindnet.es/
