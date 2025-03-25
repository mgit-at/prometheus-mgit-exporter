// Copyright (c) 2017 mgIT GmbH. All rights reserved.
// Distributed under the Apache License. See LICENSE for details.

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/anatol/luks.go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type CryptOptions struct{}

func (opts *CryptOptions) initDefault() {
}

type CryptChecker struct {
	opts CryptOptions

	promCryptSuccess             *prometheus.Desc
	promCryptManualInputRequired *prometheus.Desc
	promCryptStaticKeyFile       *prometheus.Desc
	promCryptKeySlotUsed         *prometheus.Desc
}

func NewCryptChecker(opts CryptOptions) *CryptChecker {
	opts.initDefault()
	return &CryptChecker{
		opts: opts,
		promCryptSuccess: prometheus.NewDesc(
			"mgit_crypt_success",
			"Indicates that the crypt metrics have been collected successfully",
			[]string{},
			nil),
		promCryptManualInputRequired: prometheus.NewDesc(
			"mgit_crypt_manual_input_required",
			"Indicates that a manual input is required for encryption",
			[]string{"device"},
			nil),
		promCryptStaticKeyFile: prometheus.NewDesc(
			"mgit_crypt_key_file_on_non_tmpfs",
			"Indicates that a encryption key for the specific device is stored on non tmpfs",
			[]string{"device"},
			nil),
		promCryptKeySlotUsed: prometheus.NewDesc(
			"mgit_crypt_keyslot_used",
			"Indicates that a specific luks key slot for a device is in use",
			[]string{"device", "keyslot"},
			nil),
	}
}

func (c *CryptChecker) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.promCryptSuccess
	ch <- c.promCryptManualInputRequired
	ch <- c.promCryptKeySlotUsed
	ch <- c.promCryptStaticKeyFile
}

func (c *CryptChecker) Collect(ch chan<- prometheus.Metric) {
	success := 1.0
	crypttab, err := readCrypt()
	if err != nil {
		log.Println("failed to read crypttab:", err)
		success = 0.0
	}

	for _, x := range crypttab {
		tmpFS := 0.0
		_, sourceDevice, keyfile, _ := x[0], x[1], x[2], strings.Split(x[3], ",")
		if keyfile == "none" {
			// no keyfile set - manual input required
			ch <- prometheus.MustNewConstMetric(
				c.promCryptManualInputRequired,
				prometheus.GaugeValue,
				1.0,
				sourceDevice,
			)
		} else {
			isTmpFS, err := checkTmpFS(keyfile)
			if err != nil {
				log.Printf("failed to check file type of %q: %v", keyfile, err)
				success = 0.0
			} else {
				if !isTmpFS {
					tmpFS = 1.0
				}

				ch <- prometheus.MustNewConstMetric(
					c.promCryptStaticKeyFile,
					prometheus.GaugeValue,
					tmpFS,
					sourceDevice,
				)
			}
		}

		if strings.HasPrefix(sourceDevice, "UUID=") {
			sourceDevice = "/dev/disk/by-uuid/" + strings.TrimPrefix(sourceDevice, "UUID=")
		}

		dev, err := luks.Open(sourceDevice)
		if err != nil && err.Error() == "invalid LUKS header" {
			// not a luks device.
			continue
		}
		if err != nil {
			log.Println("failed to open luks:", err)
			success = 0.0
			continue
		}

		for _, slot := range dev.Slots() {
			ch <- prometheus.MustNewConstMetric(
				c.promCryptKeySlotUsed,
				prometheus.GaugeValue,
				1.0,
				sourceDevice,
				strconv.Itoa(slot),
			)
		}
	}
	ch <- prometheus.MustNewConstMetric(
		c.promCryptSuccess,
		prometheus.GaugeValue,
		success,
	)
}

func checkTmpFS(path string) (bool, error) {
	cmdPath, err := exec.LookPath("findmnt")
	if err != nil {
		return false, errors.Wrap(err, "failed to locate findmnt command")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdPath,
		"--target", filepath.Dir(path),
		"--output", "fstype",
		"--json",
	)
	out, err := cmd.Output()
	if err != nil {
		return false, errors.Wrap(err, "failed to run findmnt")
	}

	var data struct {
		FileSystems []struct {
			Fstype string `json:"fstype"`
		} `json:"filesystems"`
	}

	if err := json.Unmarshal(out, &data); err != nil {
		return false, errors.Wrap(err, "json.Unmarshal")
	}

	if len(data.FileSystems) == 0 {
		return false, errors.New("no mount found")
	}
	if data.FileSystems[0].Fstype != "tmpfs" && data.FileSystems[0].Fstype != "devtmpfs" {
		return false, nil
	}
	return true, nil
}

func readCrypt() ([][]string, error) {
	var result [][]string
	file, err := os.Open("/etc/crypttab")
	if err != nil {
		return nil, errors.Wrap(err, "failed to open /etc/crypttab")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 4 {
			return nil, errors.Wrap(err, "failed to parse /etc/crypttab")
		}
		result = append(result, fields)
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to read /etc/crypttab")
	}
	return result, nil
}
