FROM ubuntu

RUN apt-get update
RUN apt-get upgrade

ADD inspector /opt/inspector

EXPOSE 8080
ENTRYPOINT ["/opt/inspector", "-s", "/docker/docker.sock"]
