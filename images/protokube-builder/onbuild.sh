#!/bin/bash

mkdir -p /go
export GOPATH=/go

mkdir -p /go/src/k8s.io
ln -s /src/ /go/src/k8s.io/kops

ls -lR  /go/src/k8s.io/kops/protokube/cmd/

cd /go/src/k8s.io/kops/
make protokube-gocode

mkdir -p /src/.build/artifacts/
cp /go/bin/protokube /src/.build/artifacts/
