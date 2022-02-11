FROM nicolaka/netshoot:latest
WORKDIR /
USER root:root

RUN mkdir -p /pcap && \
apk add --no-cache rsync
COPY server .
