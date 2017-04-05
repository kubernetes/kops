FROM alpine:3.5

RUN apk --no-cache add ca-certificates wget && update-ca-certificates
ADD run.sh /run.sh

CMD /run.sh