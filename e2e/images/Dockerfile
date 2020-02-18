FROM debian:buster

#RUN apt-get update && apt-get install --yes python3 && apt-get clean

RUN apt-get update
RUN apt-get install --yes curl
RUN apt-get install --yes python3 python3-nbformat python3-nbconvert python3-ipykernel
RUN apt-get install --yes python3-requests python3-xmltodict
RUN apt-get install --yes jupyter-core jupyter-nbconvert


ADD tests/ /tests
ADD lib/ /lib

ADD run.py /run.py

ENTRYPOINT [ "/run.py" ]


