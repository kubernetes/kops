#!/bin/bash

mkdir -p /go
export GOPATH=/go

mkdir -p /go/src/k8s.io
ln -s /src/ /go/src/k8s.io/kops

cd /go/src/k8s.io/kops/
make dns-controller-gocode

mkdir -p /src/.build/artifacts/
cp /go/bin/dns-controller /src/.build/artifacts/
