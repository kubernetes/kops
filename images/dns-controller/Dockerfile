FROM debian:jessie

# ca-certificates: Needed to talk to EC2 API
RUN apt-get update && apt-get install --yes ca-certificates

COPY /.build/artifacts/dns-controller /usr/bin/dns-controller

CMD /usr/bin/dns-controller

