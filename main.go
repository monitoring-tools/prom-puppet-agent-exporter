package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

const usage = `prom-puppet-agent-exporter - tool for exporting Puppet agent metrics for Prometheus

Usage:

  prom-puppet-agent-exporter [commands|flags]

The commands & flags are:

  version                   print the version to stdout
`

var (
	version   = "n/a"
	goVersion = "n/a"
	gitBranch = "n/a"
)

func main() {
	var (
		puppetYamlSummaryReportFile = flag.String("puppet.last-run-summary", "/var/lib/puppet/state/last_run_summary.yaml", "Path to puppet last_run_summary.yaml file")
		puppetYamlFullReportFile    = flag.String("puppet.last-run-report", "", "Path to puppet last_run_report.yaml file")
		puppetDisabledLockFile      = flag.String("puppet.disabled-lock", "", "Path to puppet agent_disabled.lock file")
		namespace                   = flag.String("namespace", "puppet_agent", "Namespace for metrics")
		listenAddress               = flag.String("web.listen-address", ":9001", "Address to listen on for web interface and telemetry.")
		metricsPath                 = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)

	flag.Usage = printUsage
	flag.Parse()

	handleFlags(flag.Args())

	var reportFileScraper PuppetYamlReportScraper
	if *puppetYamlFullReportFile != "" {
		reportFileScraper = NewFullReportScraper(*namespace, *puppetYamlFullReportFile, *puppetDisabledLockFile)
	} else {
		reportFileScraper = NewSummaryReportScraper(*namespace, *puppetYamlSummaryReportFile, *puppetDisabledLockFile)
	}

	buildInfoGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   *namespace,
			Subsystem:   "exporter",
			Name:        "build_info",
			Help:        "Prometheus puppet agent exporter build info in labels",
			ConstLabels: prometheus.Labels{"version": version, "goversion": goVersion, "branch": gitBranch},
		},
	)
	buildInfoGauge.Set(1)

	prometheus.MustRegister(
		NewPuppetExporter(
			*namespace,
			reportFileScraper,
		),
		buildInfoGauge,
	)

	serveHTTP(*listenAddress, *metricsPath)
}

func serveHTTP(listenAddress string, metricsEndpoint string) {
	http.Handle(metricsEndpoint, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Puppet Exporter</title></head>
			<body>
			<h1>Puppet Exporter</h1>
			<p><a href="` + metricsEndpoint + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Info("Listening on: ", formatListenAddr(listenAddress))

	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

// formatListenAddr returns formatted UNIX addr
func formatListenAddr(addr string) string {
	parts := strings.Split(addr, ":")
	if len(parts) == 2 && parts[0] == "" {
		addr = fmt.Sprintf("localhost:%s", parts[1])
	}
	return "http://" + addr
}

// handleFlags handles cli flags
func handleFlags(flags []string) {
	if len(flags) == 0 {
		return
	}

	switch flags[0] {
	case "version":
		printVersion()
	case "help":
		printUsage()
	}
}

// printUsage prints exporter usage info
func printUsage() {
	fmt.Println(usage)
	flag.PrintDefaults()

	os.Exit(0)
}

// printVersion prints exporter version
func printVersion() {
	out := version
	if out == "" {
		out = "n/a"
	}

	fmt.Println(out)
	os.Exit(0)
}
