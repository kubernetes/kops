#!/bin/bash -ex

mkdir -p /go
export GOPATH=/go

mkdir -p /go/src/k8s.io
ln -s /src/ /go/src/k8s.io/kops

ls -lR  /go/src/k8s.io/kops/protokube/cmd/

cd /go/src/k8s.io/kops/
make protokube-gocode

mkdir -p /src/.build/artifacts/
cp /go/bin/protokube /src/.build/artifacts/

# Applying channels calls out to the channels tool
make channels-gocode
cp /go/bin/channels /src/.build/artifacts/

# channels uses protokube
cd /src/.build/artifacts/
curl -O https://storage.googleapis.com/kubernetes-release/release/v1.3.7/bin/linux/amd64/kubectl
chmod +x kubectl
