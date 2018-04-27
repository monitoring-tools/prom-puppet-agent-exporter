FROM quay.io/prometheus/busybox:latest

COPY prom-puppet-agent-exporter /bin/prom-puppet-agent-exporter

EXPOSE      9001
ENTRYPOINT  [ "/bin/prom-puppet-agent-exporter" ]
