FROM debian:jessie

ARG GO_VERSION=1.8

# KOPS_GITISH: Modify to build at an explicit tag/gitish
ARG KOPS_GITISH=release

# KUBECTL_SOURCE: Change to kubernetes-dev/ci for CI
ARG KUBECTL_SOURCE=kubernetes-release/release

# KUBECTL_TRACK: Currently latest from KUBECTL_SOURCE. Change to latest-1.3.txt, etc. if desired.
ARG KUBECTL_TRACK=stable.txt

ARG KUBECTL_ARCH=linux/amd64

RUN apt-get update && \
    apt-get install -y vim ca-certificates openssh-client curl jq git gcc make && \
    curl -L https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz | tar zx -C /usr/local && \
    export PATH=$PATH:/usr/local/go/bin && \
    mkdir -p /go && \
    export GOPATH=/go && \
    go get -d k8s.io/kops && \
    cd ${GOPATH}/src/k8s.io/kops/ && \
    git checkout ${KOPS_GITISH} && \
    make && \
    mv ${GOPATH}/bin/kops /usr/bin/. && \
    KUBECTL_VERSION=$(curl -SsL --retry 5 "https://storage.googleapis.com/${KUBECTL_SOURCE}/${KUBECTL_TRACK}") && \
    curl -SsL --retry 5 "https://storage.googleapis.com/${KUBECTL_SOURCE}/${KUBECTL_VERSION}/bin/${KUBECTL_ARCH}/kubectl" > /usr/bin/kubectl && \
    chmod +x /usr/bin/kubectl &&\
    rm -rf /go /usr/local/go && \
    apt-get autoremove --purge -y curl jq git gcc make && \
    rm -rf /var/lib/apt/lists/* && \
    echo "=== Built kops at ${KOPS_GITISH}, fetched kubectl ${KUBECTL_VERSION} ==="

ENTRYPOINT ["/usr/bin/kops"]