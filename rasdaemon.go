// Copyright (c) 2017 mgIT GmbH. All rights reserved.
// Distributed under the Apache License. See LICENSE for details.

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

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

	promRasdaemonMCERecordSize *prometheus.Desc
	promRasdaemonMCEventSize   *prometheus.Desc
}

func NewRasdaemonChecker(opts RasDaemonOptions) (*RasdaemonChecker, error) {
	if opts.Path == "" {
		opts.Path = "/var/lib/rasdaemon/ras-mc_event.db"
	}

	if _, err := os.Stat(opts.Path); errors.Is(err, os.ErrNotExist) {
		return nil, errors.Wrapf(err, "file %q does not exist", opts.Path)
	}

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro", opts.Path))
	if err != nil {
		return nil, errors.Wrap(err, "sql.Open")
	}

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "db.Ping")
	}

	return &RasdaemonChecker{
		opts: opts,
		db:   db,
		promRasdaemonMCERecordSize: prometheus.NewDesc(
			"rasdaemon_mce_record_total",
			"size of the rasdaemon mce_records",
			[]string{"bank", "bank_name", "action_required"}, nil),
		promRasdaemonMCEventSize: prometheus.NewDesc(
			"rasdaemon_mc_event_total",
			"size of the rasdaemon mc-event log events",
			[]string{"err_type"}, nil),
	}, nil
}

func (c *RasdaemonChecker) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.promRasdaemonMCERecordSize
	ch <- c.promRasdaemonMCEventSize
}

func (c *RasdaemonChecker) CollectRasdaemonMCERecordSize(ch chan<- prometheus.Metric) {
	//nolint:godox
	// Todo: This could break when rasdaemon is updated.
	// See: https://pagure.io/rasdaemon/blob/master/f/mce-amd.c
	rows, err := c.db.Query(`
		select bank, bank_name, (case when error_msg like '% no action required.' then 'no' else 'yes' end) as action_required, count(id)
		from mce_record group by bank, bank_name, action_required;
	`)
	if err != nil {
		log.Println("failed to query mce_record:", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var size int
		var bank int
		var bankName string
		var actionRequired string

		if err := rows.Scan(&bank, &bankName, &actionRequired, &size); err != nil {
			log.Println("sql.Scan:", err)
			continue
		}

		// Trim unnecessary mentioning of bank in bank_name - example: bank = 18, bank_name = Unified Memory Controller (bank=18)
		bankName = strings.TrimSuffix(bankName, fmt.Sprintf(" (bank=%d)", bank))
		ch <- prometheus.MustNewConstMetric(
			c.promRasdaemonMCERecordSize,
			prometheus.GaugeValue,
			float64(size),
			strconv.Itoa(bank),
			bankName,
			actionRequired,
		)
	}
	if err := rows.Err(); err != nil {
		log.Println("sql.Next:", err)
	}
}

func (c *RasdaemonChecker) CollectRasdaemonMCEventSize(ch chan<- prometheus.Metric) {
	// There are exactly 4 error types in mc_events: Corrected, Uncorrected, Fatal and Info.
	// See: https://pagure.io/rasdaemon/blob/master/f/ras-mc-handler.c
	rows, err := c.db.Query("select err_type, count(id) from mc_event group by err_type")
	if err != nil {
		log.Println("failed to query mc_event:", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var size int
		var errType string

		if err := rows.Scan(&errType, &size); err != nil {
			log.Println("sql.Scan:", err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.promRasdaemonMCEventSize,
			prometheus.GaugeValue,
			float64(size),
			errType,
		)
	}
	if err := rows.Err(); err != nil {
		log.Println("sql.Next:", err)
	}
}

func (c *RasdaemonChecker) Collect(ch chan<- prometheus.Metric) {
	c.CollectRasdaemonMCERecordSize(ch)
	c.CollectRasdaemonMCEventSize(ch)
}

func (c *RasdaemonChecker) Close() error {
	if err := c.db.Close(); err != nil {
		return errors.Wrap(err, "sql.Close")
	}
	return nil
}
