# Copyright 2019 The Kubernetes Authors.
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

FROM alpine:3.8

ARG GO_VERSION=1.13.4

# KOPS_GITISH: Modify to build at an explicit tag/gitish
ARG KOPS_GITISH=release

# KUBECTL_SOURCE: Change to kubernetes-dev/ci for CI
ARG KUBECTL_SOURCE=kubernetes-release/release

# KUBECTL_TRACK: Currently latest from KUBECTL_SOURCE. Change to latest-1.3.txt, etc. if desired.
ARG KUBECTL_TRACK=stable.txt

ARG KUBECTL_ARCH=linux/amd64

RUN set -ex && \
        apk add --no-cache --virtual build-dependencies curl jq git bash gcc musl-dev openssl go make && \
        apk add  --no-cache vim ca-certificates &&\
        \
        export GOROOT_BOOTSTRAP="$(go env GOROOT)" && \
        curl -L https://golang.org/dl/go${GO_VERSION}.src.tar.gz | tar zx -C /usr/local && \
        cd /usr/local/go/src && \
        ./make.bash && \
        mkdir -p /go && \
        export GOPATH=/go && \
        \
        go get -d k8s.io/kops && \
        cd ${GOPATH}/src/k8s.io/kops/ && \
        git checkout ${KOPS_GITISH} && \
        make SHASUMCMD=0 && \
        mv ${GOPATH}/bin/kops /usr/bin/. && \
        \
        KUBECTL_VERSION=$(curl -SsL --retry 5 "https://storage.googleapis.com/${KUBECTL_SOURCE}/${KUBECTL_TRACK}") && \
        curl -SsL --retry 5 "https://storage.googleapis.com/${KUBECTL_SOURCE}/${KUBECTL_VERSION}/bin/${KUBECTL_ARCH}/kubectl" > /usr/bin/kubectl && \
        chmod +x /usr/bin/kubectl &&\
        \
        rm -rf /go /usr/local/go && \
        apk del build-dependencies && \
        rm -rf /var/cache/apk/* && \
        \
        echo "=== Built kops at ${KOPS_GITISH}, fetched kubectl ${KUBECTL_VERSION} ==="

ENTRYPOINT ["/usr/bin/kops"]
