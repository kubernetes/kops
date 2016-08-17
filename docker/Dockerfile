FROM golang:1.6-alpine

# KOPS_GITISH: Modify to build at an explicit tag/gitish
ARG KOPS_GITISH=HEAD

# KUBECTL_SOURCE: Change to kubernetes-dev/ci for CI
ARG KUBECTL_SOURCE=kubernetes-release/release

# KUBECTL_TRACK: Currently latest from KUBECTL_SOURCE. Change to latest-1.3.txt, etc. if desired.
ARG KUBECTL_TRACK=latest.txt

ARG KUBECTL_ARCH=linux/amd64

RUN apk add --no-cache --update build-base curl git mercurial  --virtual .kops-deps && \
    go get -d k8s.io/kops && \
    cd "${GOPATH}/src/k8s.io/kops" && git fetch origin && git reset --hard "${KOPS_GITISH}" && git clean -xdf && go get -d . && \
    make && \
    MODELS=$(readlink "${GOPATH}/bin/models") && rm "${GOPATH}/bin/models" && mv "${MODELS}" "${GOPATH}/bin/models" && \
    GITISH=$(git describe --always) && \
    KUBECTL_VERSION=${KUBECTL_VERSION:-$(curl -SsL --retry 5 "https://storage.googleapis.com/${KUBECTL_SOURCE}/${KUBECTL_TRACK}")} && \
    echo "=== Fetching kubectl ${KUBECTL_VERSION} ===" && \
    curl -SsL --retry 5 "https://storage.googleapis.com/${KUBECTL_SOURCE}/${KUBECTL_VERSION}/${KUBECTL_ARCH}/kubectl" > /usr/local/bin/kubectl && \
    chmod +x /usr/local/bin/kubectl && \
    /usr/local/bin/kubectl version --client && \
    cd / && rm -rf "${GOPATH}/src" && rm -rf "${GOPATH}/pkg" && rm -rf "/usr/local/go" && apk del .kops-deps && \
    echo "=== Built kops at ${GITISH}, fetched kubectl ${KUBECTL_VERSION} ==="

CMD "/go/bin/kops"
