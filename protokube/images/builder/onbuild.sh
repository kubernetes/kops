#!/bin/bash

mkdir -p /go
export GOPATH=/go

mkdir -p /go/src/k8s.io/kops
ln -s /src/ /go/src/k8s.io/kops/protokube

ls -lR  /go/src/k8s.io/kops/protokube/cmd/

cd /go/src/k8s.io/kops/protokube/
make gocode

mkdir -p /src/.build/artifacts/
cp /go/bin/protokube /src/.build/artifacts/
