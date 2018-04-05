#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#Build the kops.exe binary on windows

$VERSION = '0.1.0-winhack'
$GOPATH = (go env -json | ConvertFrom-Json).GOPATH
$BUILD = $GOPATH + '\src\k8s.io\kops\.build'
$DIST = $BUILD + "\dist"
New-Item -Force -ItemType Directory -Path $DIST
$GITSHA = Invoke-Command {cd "$GOPATH/src/k8s.io/kops"; git describe --always }
$env:GOOS='windows'
$Env:GOARCH='amd64'
go-bindata -o upup/models/bindata.go -pkg models -ignore="\\.DS_Store" -ignore="bindata\\.go" -ignore="vfs\\.go" -prefix upup/models/ upup/models/...
go-bindata -o federation/model/bindata.go -pkg model -ignore="\\.DS_Store" -ignore="bindata\\.go" -ignore="vfs\\.go" -prefix federation/model/ federation/model/...
go build -installsuffix cgo -o "$($DIST)/windows/amd64/kops.exe" -ldflags="-s -w -X k8s.io/kops.Version=$($VERSION) -X k8s.io/kops.GitVersion=$($GITSHA)" k8s.io/kops/cmd/kops