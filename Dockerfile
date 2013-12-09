FROM ubuntu

ADD inspector /opt/inspector

EXPOSE 8080
ENTRYPOINT ["/opt/inspector", "-s", "/docker/docker.sock"]
