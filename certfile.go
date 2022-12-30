// Copyright (c) 2017 mgIT GmbH. All rights reserved.
// Distributed under the Apache License. See LICENSE for details.

package main

import (
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"path/filepath"
	"strings"

	zglob "github.com/mattn/go-zglob"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type CertFileOptions struct {
	Globs         []string `json:"globs"`
	ExcludeSystem bool     `json:"exclude_system"`
}

type CertFileChecker struct {
	opts CertFileOptions

	promCertExpires *prometheus.Desc
}

func NewCertFileChecker(opts CertFileOptions) *CertFileChecker {
	return &CertFileChecker{
		opts: opts,
		promCertExpires: prometheus.NewDesc(
			"certfile_expires",
			"Time when the certificate will expire",
			[]string{"file", "cn"},
			nil),
	}
}

func (c *CertFileChecker) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.promCertExpires
}

func (c *CertFileChecker) Collect(ch chan<- prometheus.Metric) {
	for _, glob := range c.opts.Globs {
		matches, err := zglob.GlobFollowSymlinks(glob)
		if err != nil {
			log.Printf("failed to glob %q: %v", glob, err)
			continue
		}
		for _, m := range matches {
			if c.opts.ExcludeSystem && strings.HasPrefix(m, "/etc/ssl/certs/") {
				continue
			}
			if err := c.collectCert(m, ch); err != nil {
				log.Printf("failed to check %q: %v", m, err)
			}
		}
	}
}

func (c *CertFileChecker) collectCert(p string, ch chan<- prometheus.Metric) error {
	if !filepath.IsAbs(p) {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return errors.Wrap(err, "failed to determine absolute path")
		}
		p = absPath
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return errors.Wrapf(err, "failed to read file")
	}

	block, _ := pem.Decode(data)

	if block.Type != "CERTIFICATE" {
		return nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return errors.Wrap(err, "failed to parse certificate")
	}

	if cert.NotAfter.IsZero() {
		return nil
	}

	ch <- prometheus.MustNewConstMetric(
		c.promCertExpires,
		prometheus.GaugeValue,
		float64(cert.NotAfter.Unix()),
		p,
		cert.Subject.CommonName,
	)

	return nil
}
