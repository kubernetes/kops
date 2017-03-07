FROM golang:1.8-alpine
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
ADD .build/dist/linux/amd64/kops-server /go/bin/kops-server
ENTRYPOINT "kops-server"