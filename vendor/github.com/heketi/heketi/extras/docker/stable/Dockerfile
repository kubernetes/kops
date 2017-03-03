# set author and base
FROM centos:centos7
MAINTAINER Luis Pabón <lpabon@redhat.com>

LABEL version="1.0"
LABEL description="Centos 7 docker image for Heketi"

RUN yum --setopt=tsflags=nodocs -q -y update; yum clean all;
RUN yum --setopt=tsflags=nodocs -q -y install epel-release && \
    yum --setopt=tsflags=nodocs -q -y install heketi && \
    yum -y autoremove && \
    yum -y clean all

# post install config and volume setup
VOLUME /etc/heketi
VOLUME /var/lib/heketi

# expose port, set user and set entrypoint with config option
ENTRYPOINT ["/usr/bin/heketi"]
EXPOSE 8080

CMD ["-config=/etc/heketi/heketi.json"]
