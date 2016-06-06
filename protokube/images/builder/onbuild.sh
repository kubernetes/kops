#!/bin/bash

mkdir -p /go
export GOPATH=/go

mkdir -p /go/src/k8s.io/kube-deploy
ln -s /src/ /go/src/k8s.io/kube-deploy/protokube

ls -lR  /go/src/k8s.io/kube-deploy/protokube/cmd/

cd /go/src/k8s.io/kube-deploy/protokube/
make gocode

mkdir -p /src/.build/artifacts/
cp /go/bin/protokube /src/.build/artifacts/
