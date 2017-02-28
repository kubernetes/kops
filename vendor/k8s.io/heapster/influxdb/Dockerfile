FROM BASEIMAGE

COPY influxd /usr/bin/
COPY config.toml /etc/

ENTRYPOINT ["influxd", "--config", "/etc/config.toml"]
