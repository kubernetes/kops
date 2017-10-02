#!/usr/bin/env bash

# NOTE: Do not copy this file unless you need to use apiserver-builder at HEAD.
# Otherwise, download the pre-built apiserver-builder tar release from
# https://github.com/kubernetes-incubator/apiserver-builder/releases instead.

CMD_VENDOR="../cmd/vendor"

# Install generators from other repos
if [ ! -f bin/client-gen ] ; then
    go build -o bin/client-gen $CMD_VENDOR/k8s.io/kubernetes/cmd/libs/go2idl/client-gen
fi
if [ ! -f bin/bin/conversion-gen ] ; then
    go build -o bin/conversion-gen $CMD_VENDOR/k8s.io/kubernetes/cmd/libs/go2idl/conversion-gen
fi
if [ ! -f bin/bin/deepcopy-gen ] ; then
    go build -o bin/deepcopy-gen $CMD_VENDOR/k8s.io/kubernetes/cmd/libs/go2idl/deepcopy-gen
fi
if [ ! -f bin/openapi-gen ] ; then
    go build -o bin/openapi-gen $CMD_VENDOR/k8s.io/kubernetes/cmd/libs/go2idl/openapi-gen
fi
if [ ! -f bin/defaulter-gen ] ; then
    go build -o bin/defaulter-gen $CMD_VENDOR/k8s.io/kubernetes/cmd/libs/go2idl/defaulter-gen
fi
if [ ! -f bin/lister-gen ] ; then
    go build -o bin/lister-gen $CMD_VENDOR/k8s.io/kubernetes/cmd/libs/go2idl/lister-gen
fi
if [ ! -f bin/informer-gen ] ; then
    go build -o bin/informer-gen $CMD_VENDOR/k8s.io/kubernetes/cmd/libs/go2idl/informer-gen
fi
if [ ! -f bin/gen-apidocs ] ; then
    go build -o bin/gen-apidocs $CMD_VENDOR/github.com/kubernetes-incubator/reference-docs/gen-apidocs
fi

# Install generators from this repo
go build -o bin/apiserver-boot ../cmd/apiserver-boot/main.go
go build -o bin/apiregister-gen ../cmd/apiregister-gen/main.go
