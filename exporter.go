package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// PuppetExporter stats exporter
type PuppetExporter struct {
	namespace     string
	reportScraper PuppetYamlReportScraper

	scrapesSummary prometheus.Summary
	failedScrapes  prometheus.Counter
}

const subsystem = "exporter"

// NewPuppetExporter creates puppet stats exporter
func NewPuppetExporter(
	namespace string,
	reportScraper PuppetYamlReportScraper,
) *PuppetExporter {
	scrapesSummary := prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "scrape_duration_seconds",
		Help:      "The scrapes durations summary and total count.",
	})

	failedScrapes := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "scrapes_failed",
		Help:      "Current failed puppet scrapes.",
	})

	exporter := &PuppetExporter{
		namespace:      namespace,
		reportScraper:  reportScraper,
		scrapesSummary: scrapesSummary,
		failedScrapes:  failedScrapes,
	}

	return exporter
}

// Describe implements prometheus.Collector interface
func (exp *PuppetExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- exp.scrapesSummary.Desc()
	ch <- exp.failedScrapes.Desc()
}

// Collect implements prometheus.Collector interface
func (exp *PuppetExporter) Collect(ch chan<- prometheus.Metric) {
	if err := exp.metrics(ch); err != nil {
		exp.failedScrapes.Inc()
	}

	ch <- exp.scrapesSummary
	ch <- exp.failedScrapes
}

func (exp *PuppetExporter) metrics(ch chan<- prometheus.Metric) error {
	now := time.Now()
	defer exp.scrapesSummary.Observe(time.Since(now).Seconds())
	return exp.reportScraper.CollectMetrics(ch)
}
