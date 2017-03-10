// Copyright (c) 2017 mgIT GmbH. All rights reserved.
// Distributed under the Apache License. See LICENSE for details.

package main

import (
	"log"
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

type MCELogOptions struct {
	Path string `json:"path"`
}

type MCELogChecker struct {
	opts MCELogOptions

	promMCELogSize *prometheus.Desc
}

func NewMCELogChecker(opts MCELogOptions) *MCELogChecker {
	if opts.Path == "" {
		opts.Path = "/var/log/mcelog"
	}

	return &MCELogChecker{
		opts: opts,
		promMCELogSize: prometheus.NewDesc(
			"mcelog_size_bytes",
			"size of the machine exception log",
			[]string{"file"},
			nil),
	}
}

func (c *MCELogChecker) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.promMCELogSize
}

func (c *MCELogChecker) Collect(ch chan<- prometheus.Metric) {
	var size int64

	info, err := os.Stat(c.opts.Path)
	if err != nil {
		if os.IsNotExist(err) {
			size = 0
		} else {
			log.Println("failed to get mcelog size:", err)
			return
		}
	} else {
		size = info.Size()
	}

	ch <- prometheus.MustNewConstMetric(
		c.promMCELogSize,
		prometheus.GaugeValue,
		float64(size),
		c.opts.Path,
	)
}
