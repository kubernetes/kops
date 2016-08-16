FROM debian:jessie

RUN apt-get update && apt-get install --yes curl python-pip openssh-client
RUN pip install awscli

RUN curl https://sdk.cloud.google.com | bash

RUN curl https://storage.googleapis.com/golang/go1.6.3.linux-amd64.tar.gz | tar -C /usr/local -xzf -

ENV PATH /root/google-cloud-sdk/bin:/usr/local/go/bin:$PATH

ADD runtests.sh /

ADD conf /conf
