#!/bin/sh

GRAFANA_CREDENTIALS=admin:admin
GRAFANA_HOST_PORT=grafana:3000
PROMETHEUS_HOST_PORT=prometheus:9090

/root/wait-for.sh $GRAFANA_HOST_PORT -- echo "Provisioning Grafana @ $GRAFANA_HOST_PORT" && \
curl -q -u $GRAFANA_CREDENTIALS -XDELETE http://$GRAFANA_HOST_PORT/api/datasources/name/prometheus  && \
curl -q -u $GRAFANA_CREDENTIALS -XPOST -d '{"name":"prometheus","type":"prometheus","url":"http://'$PROMETHEUS_HOST_PORT'","access":"proxy","basicAuth":false}' -H 'Content-Type: application/json' http://$GRAFANA_HOST_PORT/api/datasources
