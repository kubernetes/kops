#!/usr/bin/env bash

#make nodeup-arm64
make nodeup-amd64
#make protokube-arm64
make protokube-amd64
#sha256sum .build/dist/linux/amd64/nodeup > ./build/dist/linux/amd64/hash-nodeup
#sha256sum .build/dist/linux/arm64/nodeup > ./build/dist/linux/arm64/hash-nodeup
rclone sync .build/dist/ scaleway:kops-state-store-test/dist/ -P
#mc cp -r .build/dist/ s3://kops-state-store-test/dist/