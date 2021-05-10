// Copyright (c) 2017 mgIT GmbH. All rights reserved.
// Distributed under the Apache License. See LICENSE for details.

package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	_ "modernc.org/sqlite"
)

type RasDaemonOptions struct {
	Path string `json:"path"`
}

type RasdaemonChecker struct {
	opts RasDaemonOptions
	db   *sql.DB

	promRasdaemonSize *prometheus.Desc
}

func NewRasdaemonChecker(opts RasDaemonOptions) (*RasdaemonChecker, error) {
	if opts.Path == "" {
		opts.Path = "/var/lib/rasdaemon/ras-mc_event.db"
	}

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro", opts.Path))
	if err != nil {
		return nil, errors.Wrap(err, "sql.Open")
	}

	return &RasdaemonChecker{
		opts: opts,
		db:   db,
		promRasdaemonSize: prometheus.NewDesc(
			"rasdaemon_entries_total",
			"size of the rasdaemon mc-event log",
			[]string{"event"}, nil),
	}, nil
}

func (c *RasdaemonChecker) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.promRasdaemonSize
}

func (c *RasdaemonChecker) Collect(ch chan<- prometheus.Metric) {
	rows, err := c.db.Query("select bank_name, count(id) from mce_record group by bank_name")
	if err != nil {
		log.Println("failed to query mce_record:", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var size int
		var event string

		if err := rows.Scan(&event, &size); err != nil {
			log.Println("sql.Scan:", err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.promRasdaemonSize,
			prometheus.GaugeValue,
			float64(size),
			event,
		)
	}
	if err := rows.Err(); err != nil {
		log.Println("sql.Next:", err)
	}
}

func (c *RasdaemonChecker) Close() error {
	if err := c.db.Close(); err != nil {
		return errors.Wrap(err, "sql.Close")
	}
	return nil
}
