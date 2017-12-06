// Copyright (c) 2017 mgIT GmbH. All rights reserved.
// Distributed under the Apache License. See LICENSE for details.

package main

import (
	"bytes"
	"context"
	"log"
	"os/exec"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type PTHeartbeatOptions struct {
	Database     string `json:"database"`
	Table        string `json:"table"`
	DefaultsFile string `json:"defaultsFile"`
	MasterID     int    `json:"masterId"`
}

func (opts *PTHeartbeatOptions) initDefault() {
	if opts.Database == "" {
		opts.Database = "system"
	}
	if opts.Table == "" {
		opts.Table = "pt_heartbeat"
	}
	if opts.DefaultsFile == "" {
		opts.DefaultsFile = "/etc/mysql/debian.cnf"
	}
}

type PTHeartbeatChecker struct {
	opts PTHeartbeatOptions

	promLag     *prometheus.Desc
	promSuccess *prometheus.Desc
}

func NewPTHeartbeatChecker(opts PTHeartbeatOptions) *PTHeartbeatChecker {
	opts.initDefault()
	return &PTHeartbeatChecker{
		opts: opts,
		promLag: prometheus.NewDesc(
			"ptheartbeat_lag_seconds",
			"MySQL replication lag measured by pt-heartbeat",
			[]string{"database", "table", "master"},
			nil),
		promSuccess: prometheus.NewDesc(
			"ptheartbeat_success",
			"Indicates that the replication lag has been collected successfully",
			[]string{"database", "table", "master"},
			nil),
	}
}

func (c *PTHeartbeatChecker) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.promSuccess
	ch <- c.promLag
}

func (c *PTHeartbeatChecker) Collect(ch chan<- prometheus.Metric) {
	success := 1.0
	lag, err := c.collectLag()
	if err != nil {
		log.Println("failed to collect ptheartbeat:", err)
		if exitErr, ok := errors.Cause(err).(*exec.ExitError); ok {
			if len(exitErr.Stderr) > 0 {
				log.Printf("additional output (stderr): %s\n", exitErr.Stderr)
			}
		}
		success = 0.0
	}
	ch <- prometheus.MustNewConstMetric(
		c.promSuccess,
		prometheus.GaugeValue,
		success,
		c.opts.Database, c.opts.Table, strconv.Itoa(c.opts.MasterID),
	)
	if err == nil {
		ch <- prometheus.MustNewConstMetric(
			c.promLag,
			prometheus.GaugeValue,
			lag,
			c.opts.Database, c.opts.Table, strconv.Itoa(c.opts.MasterID),
		)
	}
}

func (c *PTHeartbeatChecker) collectLag() (float64, error) {
	cmdPath, err := exec.LookPath("pt-heartbeat")
	if err != nil {
		return 0, errors.Wrap(err, "failed to locate pt-heartbeat command")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdPath,
		"--check",
		"--database", c.opts.Database,
		"--table", c.opts.Table,
		"--defaults-file", c.opts.DefaultsFile,
		"--master-server-id", strconv.Itoa(c.opts.MasterID),
		"--noinsert-heartbeat-row",
		"--utc",
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, errors.Wrap(err, "failed to run pt-heartbeat")
	}
	out = bytes.TrimSpace(out)

	lag, err := strconv.ParseFloat(string(out), 64)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse pt-heartbeat output")
	}

	return lag, nil
}
