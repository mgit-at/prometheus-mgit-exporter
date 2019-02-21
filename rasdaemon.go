// Copyright (c) 2017 mgIT GmbH. All rights reserved.
// Distributed under the Apache License. See LICENSE for details.

package main

import (
	"bytes"
	"log"
	"os/exec"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type RasDaemonOptions struct {
	Path string `json:"path"`
}

type RasdaemonChecker struct {
	opts RasDaemonOptions

	promRasdaemonSize *prometheus.Desc
}

func NewRasdaemonChecker(opts RasDaemonOptions) *RasdaemonChecker {
	if opts.Path == "" {
		opts.Path = "/var/lib/rasdaemon/ras-mc_event.db"
	}

	return &RasdaemonChecker{
		opts: opts,
		promRasdaemonSize: prometheus.NewDesc(
			"rasdaemon_entries_total",
			"size of the rasdaemon mc-event log",
			nil, nil),
	}
}

func (c *RasdaemonChecker) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.promRasdaemonSize
}

func (c *RasdaemonChecker) Collect(ch chan<- prometheus.Metric) {
	cmd := exec.Command("sqlite3",
		c.opts.Path,
		"select count(*) from mce_record;",
	)

	output, err := cmd.Output()
	if err != nil {
		log.Println("failed to run sqlite3:", err)
	}

	output = bytes.TrimSpace(output)

	size, err := strconv.ParseInt(string(output), 10, 64)
	if err != nil {
		log.Println("failed to parse sqlite3 output")
	}

	ch <- prometheus.MustNewConstMetric(
		c.promRasdaemonSize,
		prometheus.GaugeValue,
		float64(size),
	)
}
