FROM debian:jessie

# Install packages:
#  curl (to download glide & golang)
#  git, mercurial (for go get)
RUN apt-get update && apt-get install --yes curl mercurial git gcc make

# Install golang
RUN curl -L https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz | tar zx -C /usr/local
ENV PATH $PATH:/usr/local/go/bin
# Install glide
RUN curl -L https://github.com/Masterminds/glide/releases/download/0.10.2/glide-0.10.2-linux-amd64.tar.gz | tar zx --strip-components 1 -C /usr/bin

COPY images/builder/onbuild.sh /onbuild.sh
