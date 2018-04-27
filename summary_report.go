package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

// NewSummaryReportScraper creates new scraper for puppet agent summary report file
func NewSummaryReportScraper(namespace, reportFilename, disableLockFilename string) PuppetYamlReportScraper {
	return &summaryReportScraper{
		newPuppetReportScraper(namespace, reportFilename, disableLockFilename),
	}
}

type summaryReportScraper struct {
	*reportScraper
}

func (r *summaryReportScraper) CollectMetrics(ch chan<- prometheus.Metric) error {
	return r.collectMetrics(ch, r)
}

func (r *summaryReportScraper) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v struct {
		Version struct {
			Puppet string
			Config float64
		}
	}
	if err := unmarshal(&v); err != nil {
		return err
	}

	r.reportScraper.setPuppetVersion(v.Version.Puppet)
	r.reportScraper.setConfigTimestamp(v.Version.Config)

	var objmap map[string]gaugeValueMap
	unmarshal(&objmap)
	delete(objmap, "version")

	for sectionName, reportSection := range objmap {
		for metricName, metricValue := range reportSection {
			r.reportScraper.setMetricValue(sectionName, metricName, metricValue)
		}
	}

	return nil
}
