FROM debian:jessie

# Install packages:
#  curl (to download golang)
#  gcc & make (for building)
RUN apt-get update && apt-get install --yes curl gcc make

# Install golang
RUN curl -L https://storage.googleapis.com/golang/go1.6.3.linux-amd64.tar.gz | tar zx -C /usr/local
ENV PATH $PATH:/usr/local/go/bin

COPY images/dns-controller-builder/onbuild.sh /onbuild.sh
