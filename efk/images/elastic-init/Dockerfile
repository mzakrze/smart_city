FROM ubuntu

RUN apt-get update && \
apt-get install -y curl

WORKDIR /usr/local

COPY elasticsearch_init.sh .

ENTRYPOINT ["bash", "elasticsearch_init.sh" ]

