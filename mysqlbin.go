package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type MySQLBinOptions struct {
	Path string `json:"path"`
}

func (opts *MySQLBinOptions) initDefault() {
	if opts.Path == "" {
		opts.Path = "/var/log/mysql"
	}
}

type MySQLBinChecker struct {
	opts MySQLBinOptions

	promBinlogOk *prometheus.Desc
}

func NewMySQLBinChecker(opts MySQLBinOptions) *MySQLBinChecker {
	opts.initDefault()
	return &MySQLBinChecker{
		opts: opts,
		promBinlogOk: prometheus.NewDesc(
			"mysql_binlog_ok",
			"are all binlog files from the index available?",
			[]string{},
			nil),
	}
}

func (c *MySQLBinChecker) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.promBinlogOk
}

func (c *MySQLBinChecker) Collect(ch chan<- prometheus.Metric) {
	ok, err := c.checkBinlogs()
	if err != nil {
		log.Println("invalid binlog files detected:", err)
	}
	value := 0.0
	if ok {
		value = 1.0
	}
	ch <- prometheus.MustNewConstMetric(
		c.promBinlogOk,
		prometheus.GaugeValue,
		value,
	)
}

func (c *MySQLBinChecker) checkBinlogs() (bool, error) {
	indexFile, err := os.Open(filepath.Join(c.opts.Path, "mysql-bin.index"))
	if err != nil {
		return false, errors.Wrap(err, "failed to open index file")
	}
	defer indexFile.Close()

	scanner := bufio.NewScanner(indexFile)
	for scanner.Scan() {
		filename := strings.TrimSpace(scanner.Text())
		if filename == "" {
			continue
		}
		if !filepath.IsAbs(filename) {
			filename = filepath.Join(c.opts.Path, filename)
		}
		info, err := os.Stat(filename)
		if err != nil {
			if os.IsNotExist(err) {
				return false, errors.Wrapf(err, "missing file %q", filename)
			}
			return false, errors.Wrapf(err, "failed to stat file %q", filename)
		}
		if info.IsDir() {
			return false, fmt.Errorf("unexpected directory %q", filename)
		}
	}
	if err := scanner.Err(); err != nil {
		return false, errors.Wrap(err, "failed to scan index file")
	}
	return true, nil
}
