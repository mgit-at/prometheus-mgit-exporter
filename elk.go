// Copyright (c) 2017 mgIT GmbH. All rights reserved.
// Distributed under the Apache License. See LICENSE for details.

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type ElkOptions struct {
	Duration string `json:"duration"`
	Node     string `json:"node"`
}

type ElkChecker struct {
	opts ElkOptions

	promElkSize *prometheus.Desc
}

func NewElkChecker(opts ElkOptions) *ElkChecker {
	if opts.Duration == "" {
		opts.Duration = "170h" // 7d2h
	}
	if opts.Node == "" {
		opts.Node = "hot"
	}
	return &ElkChecker{
		opts: opts,
		promElkSize: prometheus.NewDesc(
			"elk_indices_not_moved",
			"number of elk indices that are not moved",
			[]string{},
			nil),
	}
}

func (c *ElkChecker) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.promElkSize
}

func (c *ElkChecker) Collect(ch chan<- prometheus.Metric) {
	moved, err := c.collectBadIndices()
	if err != nil {
		log.Println("failed to get bad indices", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.promElkSize,
		prometheus.GaugeValue,
		float64(moved),
	)
}

func (c *ElkChecker) collectBadIndices() (int, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequest("GET", "http://127.0.0.1:9200/_cat/shards?format=json", nil)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create request")
	}
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get shard status")
	}
	defer resp.Body.Close()

	var shards []shard
	err = json.NewDecoder(resp.Body).Decode(&shards)
	if err != nil {
		return 0, errors.Wrap(err, "failed to unmarshal shards")
	}

	return c.countBadIndices(shards)
}

func (c *ElkChecker) countBadIndices(shards []shard) (int, error) {
	dur, err := time.ParseDuration(c.opts.Duration)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse duration")
	}
	counter := 0
	for _, s := range shards {
		if s.Node != c.opts.Node {
			continue
		}
		older, err := s.isOlder(dur)
		if err != nil {
			return 0, errors.Wrap(err, "failed to count bad indices")
		}
		if older {
			counter++
		}
	}
	return counter, nil
}

type shard struct {
	Index  string `json:"index"`
	Shard  string `json:"shard"`
	Prirep string `json:"prirep"`
	State  string `json:"state"`
	Docs   string `json:"docs"`
	Store  string `json:"store"`
	IP     string `json:"ip"`
	Node   string `json:"node"`
}

func (s *shard) isOlder(dur time.Duration) (bool, error) {
	now := time.Now()
	now = now.Add(dur * -1)

	re := regexp.MustCompile(`((?:\d{4}(.|-)\d{2}(.|-)\d{2}))$`)
	d := re.FindString(s.Index)
	if d == "" {
		return false, nil
	}

	date := strings.ReplaceAll(d, ".", "-")
	const shortForm = "2006-01-02"
	shardDate, err := time.Parse(shortForm, date)
	if err != nil {
		return false, errors.Wrap(err, "failed to parse date")
	}

	if now.After(shardDate) {
		return true, nil
	}
	return false, nil
}
