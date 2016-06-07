FROM debian:jessie

# ca-certificates: Needed to talk to EC2 API
# e2fsprogs: Needed to mount / format ext4 filesytems
RUN apt-get update && apt-get install --yes ca-certificates e2fsprogs

COPY model/ /model/
COPY templates/ /templates/
COPY /.build/artifacts/protokube /usr/bin/protokube

CMD /usr/bin/protokube

