FROM alpine:3.6

RUN apk update && apk add openssh-client bash
RUN echo "StrictHostKeyChecking no" >> /etc/ssh/ssh_config

RUN echo "UserKnownHostsFile /dev/null" >> /etc/ssh/ssh_config
ADD machine-controller /bin

ENTRYPOINT ["/bin/machine-controller"]
