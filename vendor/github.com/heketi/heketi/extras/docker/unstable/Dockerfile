# set author and base
FROM fedora
MAINTAINER Luis Pabón <lpabon@redhat.com>

LABEL version="1.3.1"
LABEL description="Development build"

# let's setup all the necessary environment variables
ENV BUILD_HOME=/build
ENV GOPATH=$BUILD_HOME/golang
ENV PATH=$GOPATH/bin:$PATH
ENV HEKETIC_CONF_DIR=/etc/heketi

# install dependencies, build and cleanup
RUN mkdir $BUILD_HOME $GOPATH $HEKETI_CONF_DIR && \
    dnf -q -y install golang git && \
    dnf -q -y install make && \
    dnf -q -y clean all && \
    mkdir -p $GOPATH/src/github.com/heketi && \
    cd $GOPATH/src/github.com/heketi && \
    git clone https://github.com/heketi/heketi.git && \
    go get github.com/tools/godep && \
    cd $GOPATH/src/github.com/heketi/heketi && make && \
    cp heketi /usr/bin/heketi && \
    cp client/cli/go/heketi-cli /usr/bin/heketi-cli && \
    cd && rm -rf $BUILD_HOME && \
    dnf -q -y remove git golang make && \
    dnf -q -y autoremove && \
    dnf -q -y clean all

# post install config and volume setup
ADD ./heketi.json /etc/heketi/heketi.json
VOLUME /etc/heketi

RUN mkdir /var/lib/heketi
VOLUME /var/lib/heketi

# expose port, set user and set entrypoint with config option
ENTRYPOINT ["/usr/bin/heketi"]
EXPOSE 8080

CMD ["-config=/etc/heketi/heketi.json"]
