ARG BASE_IMAGE

FROM ${BASE_IMAGE}

ADD tests/ /tests
ADD lib/ /lib

ADD run.py /run.py

ENTRYPOINT [ "/run.py" ]


