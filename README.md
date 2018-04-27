# Prom Puppet Agent Exporter

`Prom Puppet Agent Exporter` is puppet agent report statistics exporter for Prometheus.

The `Prom Puppet Agent Exporter` reads last run summary report and exposes them for Prometheus consumption.

## Building and running

```
$ make build
make build
>> installing golang dependencies into vendor directory...
>> building binary...
>> GOOS= GOARCH= GO15VENDOREXPERIMENT=1 go build -ldflags -X main.version=0.2.0-7-gccf3901 -X main.goVersion=go version go1.8.3 darwin/amd64 -X main.gitBranch=GOLIBS-1534 -o prom-puppet-agent-exporter
$ ./prom-puppet-agent-exporter-linux -web.listen-address "localhost:9005" -web.telemetry-path "/metrics" -namespace "puppet_agent" -puppet.last-run-report="/var/lib/puppet/state/last_run_report.yaml"
```

### Other useful make commands:

The building application for current system:
```
$ make build
```

The running unit tests:
```
$ make test
```

The applying go tool to code:
```
$ make fmt
>> formatting source

$ make lint
>> linting source

$ make imports
>> fixing source imports
```

### Flags

Name                    | Default                                     | Description
----------------------- | ------------------------------------------- | -----------
web.listen-address      | localhost:9001                              | Address on which to expose metrics and web interface.
web.telemetry-path      | /metrics                                    | Path under which to expose metrics.
namespace               | puppet                                      | The namespace of metrics.
puppet.last-run-summary | /var/lib/puppet/state/last_run_summary.yaml | Path to the puppet's last run summary report.
puppet.last-run-report  |                                             | Path to the puppet's last run full report - need only if you want to get info from full report.
puppet.disabled-lock    |                                             | Path to the puppet's agent disabled lock-file.

## What's exported?
It exports statistics from standard Puppet report (https://puppet.com/blog/puppet-monitoring-how-to-monitor-success-or-failure-of-puppet-runs).
