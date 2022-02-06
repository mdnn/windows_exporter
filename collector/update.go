//go:build windows
// +build windows

package collector

import (
	"errors"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows/registry"
	"strconv"
)

func init() {
	registerCollector("update", NewUpdateCollector, "Windows Update")
}

// A UpdateCollector is a Prometheus collector for WMI metrics
type UpdateCollector struct {
	RebootRequired *prometheus.Desc
}

// NewUpdateCollector ...
func NewUpdateCollector() (Collector, error) {
	const subsystem = "update"

	return &UpdateCollector{
		RebootRequired: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "reboot"),
			"Reboot required",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *UpdateCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting update metrics:", desc, err)
		return err
	}
	return nil
}

func (c *UpdateCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	// Get current build from registry
	cvKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer cvKey.Close()

	cbKey, _, err := cvKey.GetStringValue("CurrentBuild")
	if err != nil {
		return nil, err
	}

	cb, err := strconv.Atoi(cbKey)
	if err != nil {
		return nil, err
	}

	if cb < 14393 {
		return nil, errors.New("windows version older than Server 2016 detected")
	}

	var up int
	// Get reboot status from registry
	ntKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update\RebootRequired`, registry.QUERY_VALUE)
	defer ntKey.Close()

	if err != nil {
		up = 0
	} else {
		up = 1
	}

	ch <- prometheus.MustNewConstMetric(
		c.RebootRequired,
		prometheus.GaugeValue,
		float64(up),
	)

	return nil, nil
}
