// Copyright (c) 2017 mgIT GmbH. All rights reserved.
// Distributed under the Apache License. See LICENSE for details.

package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type FsTabOptions struct{}

func (opts *FsTabOptions) initDefault() {
}

type FsTabChecker struct {
	opts FsTabOptions

	promSuccess *prometheus.Desc
	promNoFail  *prometheus.Desc
}

func NewFsTabChecker(opts FsTabOptions) *FsTabChecker {
	opts.initDefault()
	return &FsTabChecker{
		opts: opts,
		promSuccess: prometheus.NewDesc(
			"mgit_fstab_success",
			"Indicates that the fstab has been collected successfully",
			[]string{},
			nil),
		promNoFail: prometheus.NewDesc(
			"mgit_fstab_nofail",
			"Indicates that the nofail flag is set for the specific remote fs",
			[]string{"device", "mountpoint", "fstype"},
			nil),
	}
}

func (c *FsTabChecker) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.promSuccess
	ch <- c.promNoFail
}

func (c *FsTabChecker) Collect(ch chan<- prometheus.Metric) {
	fstab, err := c.readFsTab()

	success := 1.0
	if err != nil {
		log.Println("failed to collect fstab:", err)
		success = 0.0
	}

	ch <- prometheus.MustNewConstMetric(
		c.promSuccess,
		prometheus.GaugeValue,
		success,
	)

	for _, x := range fstab {
		device, mountpoint, fstype, flags := x[0], x[1], x[2], strings.Split(x[3], ",")

		if fstype != "nfs" {
			continue
		}

		nofail := 0.0
		for _, f := range flags {
			if f == "nofail" {
				nofail = 1.0
			}
		}

		ch <- prometheus.MustNewConstMetric(
			c.promNoFail,
			prometheus.GaugeValue,
			nofail,
			device, mountpoint, fstype,
		)
	}
}

func (c *FsTabChecker) readFsTab() ([][]string, error) {
	var result [][]string
	file, err := os.Open("/etc/fstab")
	if err != nil {
		return nil, errors.Wrap(err, "failed to open /etc/fstab")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 6 {
			return nil, errors.Wrap(err, "failed to parse /etc/fstab")
		}
		result = append(result, fields)
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to read /etc/fstab")
	}
	return result, nil
}
