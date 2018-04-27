package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
)

var (
	knownMetricsDescriptionsBySection = map[string]map[string]string{
		"resources": {
			"failed":            "How many resources were not successfully fixed",
			"scheduled":         "How many resources met any scheduling restrictions",
			"changed":           "How many resources were correctively changed",
			"skipped":           "How many resources were skipped, because of either tagging or scheduling restrictions",
			"total":             "The total number of resources being managed",
			"failed_to_restart": "How many resources could not be restarted",
			"restarted":         "How many resources was restarted",
			"out_of_sync":       "How many resources were out of sync",
			"corrective_change": "Number of corrected resources",
		},
		"events": {
			"failure": "How many events were failed",
			"success": "How many events were succeed",
			"total":   "Total number of events for last run",
		},
		"changes": {
			"total": "Total number of changes in the transaction",
		},
		"time": {},
	}
)

// PuppetYamlReportScraper defines a report scraper interface
type PuppetYamlReportScraper interface {
	CollectMetrics(ch chan<- prometheus.Metric) error
}

type gaugeValueMap map[string]float64

type gaugeValuesBySectionsMap map[string]gaugeValueMap

type puppetYamlReportUnmarshaler interface {
	yaml.Unmarshaler
}

type reportScraper struct {
	namespace            string
	reportFilename       string
	disabledLockFilename string

	info            map[string]string
	data            gaugeValuesBySectionsMap
	puppetVersion   string
	configTimestamp float64
}

func newPuppetReportScraper(namespace, reportFilename, disableLockFilename string) *reportScraper {
	r := &reportScraper{
		namespace:            namespace,
		reportFilename:       reportFilename,
		disabledLockFilename: disableLockFilename,
		info:                 map[string]string{},
	}

	r.setInfo("report_filename", reportFilename)

	if disableLockFilename != "" {
		r.setInfo("disabled_lock_filename", disableLockFilename)
	}

	return r
}

func (r *reportScraper) collectMetrics(ch chan<- prometheus.Metric, u puppetYamlReportUnmarshaler) error {
	if r.disabledLockFilename != "" {
		agentDisabled, disabledMessage := r.processDisabledLock()

		var disabledLockMetricValue float64
		if agentDisabled {
			disabledLockMetricValue = 1
		}

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(r.namespace, "", "disabled"),
				"Puppet state of agent disabled lock",
				[]string{"disabled_message"},
				nil,
			),
			prometheus.CounterValue,
			disabledLockMetricValue,
			[]string{disabledMessage}...,
		)
	}

	metricsBySections, err := r.processReport(u)

	if metricsBySections != nil {
		for section, metrics := range metricsBySections {
			for metricName, metricValue := range metrics {
				help, ok := knownMetricsDescriptionsBySection[section][metricName]
				if !ok {
					help = metricName
				}

				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						prometheus.BuildFQName(r.namespace, section, metricName),
						help,
						nil,
						nil,
					),
					prometheus.GaugeValue,
					metricValue,
				)
			}
		}
	}

	info := r.info

	delete(info, "error")

	if err != nil {
		info["error"] = err.Error()
	}

	info["puppet_version"] = r.puppetVersion

	var infoKeys, infoValues []string
	for k, v := range info {
		if v != "" {
			infoKeys = append(infoKeys, k)
			infoValues = append(infoValues, v)
		}
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(r.namespace, "", "info"),
			"Puppet configuration info in labels",
			infoKeys,
			nil,
		),
		prometheus.GaugeValue,
		r.configTimestamp,
		infoValues...,
	)

	return err
}

func (r *reportScraper) setMetricValue(section, name string, value float64) {
	if _, ok := r.data[section]; !ok {
		r.data[section] = make(gaugeValueMap)
	}
	r.data[section][name] = value
}

func (r *reportScraper) setInfo(key, value string) {
	r.info[key] = value
}

func (r *reportScraper) setPuppetVersion(version string) {
	r.puppetVersion = version
}

func (r *reportScraper) setConfigTimestamp(value float64) {
	r.configTimestamp = value
}

type parseError struct {
	filename string
	error
}

func (err *parseError) Error() string {
	return fmt.Sprintf("Error occured while trying to parse report file %q: %s", err.filename, err.error)
}

func (r *reportScraper) processDisabledLock() (bool, string) {
	disabledLockContent, err := readFile(r.disabledLockFilename)

	disabled := true
	switch err.(type) {
	case *notFoundError:
		disabled = false
	}

	var d struct{ DisabledMessage string `json:"disabled_message"` }
	if disabledLockContent != nil {
		if err := json.Unmarshal(disabledLockContent, &d); err != nil {
			return disabled, err.Error()
		}
	}

	return disabled, d.DisabledMessage
}

func (r *reportScraper) processReport(u puppetYamlReportUnmarshaler) (gaugeValuesBySectionsMap, error) {
	reportContent, err := readFile(r.reportFilename)

	if err != nil {
		return nil, err
	}

	r.data = make(gaugeValuesBySectionsMap)

	err = yaml.Unmarshal(reportContent, u)
	if err != nil {
		return nil, &parseError{r.reportFilename, err}
	}

	return r.data, nil
}

type readError struct {
	filename string
	error
}

type notFoundError struct {
	*readError
}

func (err *readError) Error() string {
	return fmt.Sprintf("Error occured while trying to read report file %q: %s", err.filename, err.error)
}

func readFile(filename string) ([]byte, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, &notFoundError{&readError{filename, err}}
	}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, &readError{filename, err}
	}

	return content, nil
}
