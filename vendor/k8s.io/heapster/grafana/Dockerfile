FROM BASEIMAGE

ADD grafana.tar /
COPY dashboards /dashboards
COPY run.sh /
COPY setup_grafana /usr/bin/

ENTRYPOINT ["/run.sh"]
