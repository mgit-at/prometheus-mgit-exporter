// Copyright (c) 2017 mgIT GmbH. All rights reserved.
// Distributed under the Apache License. See LICENSE for details.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	Listen   string `json:"listen"`
	CertFile struct {
		Enable bool `json:"enable"`
		CertFileOptions
	} `json:"certfile"`
	MCELog struct {
		Enable bool `json:"enable"`
		MCELogOptions
	} `json:"mcelog"`
	PTHeartbeat struct {
		Enable bool `json:"enable"`
		PTHeartbeatOptions
	} `json:"ptheartbeat"`
	FsTab struct {
		Enable bool `json:"enable"`
		FsTabOptions
	}
}

func run() error {
	var (
		flagConfig = flag.String("config", "config.json", "configuration file")
	)
	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		return fmt.Errorf("invalid number of arguments")
	}

	cfgFile, err := os.Open(*flagConfig)
	if err != nil {
		return fmt.Errorf("failed to open config %q: %v", *flagConfig, err)
	}
	defer cfgFile.Close()

	var cfg Config
	if err := json.NewDecoder(cfgFile).Decode(&cfg); err != nil {
		return fmt.Errorf("failed to decode config %q: %v", *flagConfig, err)
	}

	if cfg.CertFile.Enable {
		log.Println("enabling certfile checker")
		c := NewCertFileChecker(cfg.CertFile.CertFileOptions)
		if err := prometheus.Register(c); err != nil {
			return fmt.Errorf("failed to register certfile checker: %v", err)
		}
	}

	if cfg.MCELog.Enable {
		log.Println("enabling mcelog checker")
		c := NewMCELogChecker(cfg.MCELog.MCELogOptions)
		if err := prometheus.Register(c); err != nil {
			return fmt.Errorf("failed to register mcelog checker: %v", err)
		}
	}

	if cfg.PTHeartbeat.Enable {
		log.Println("enabling ptheartbeat checker")
		c := NewPTHeartbeatChecker(cfg.PTHeartbeat.PTHeartbeatOptions)
		if err := prometheus.Register(c); err != nil {
			return fmt.Errorf("failed to register ptheartbeat checker: %v", err)
		}
	}

	if cfg.FsTab.Enable {
		log.Println("enabling fstab checker")
		c := NewFsTabChecker(cfg.FsTab.FsTabOptions)
		if err := prometheus.Register(c); err != nil {
			return fmt.Errorf("failed to register fstab checker: %v", err)
		}
	}

	if cfg.Listen == "" {
		cfg.Listen = ":9328"
	}
	listen, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		return fmt.Errorf("failed to listen at %q: %v", cfg.Listen, err)
	}
	defer listen.Close()
	log.Println("listening on", listen.Addr())

	http.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  5 * time.Minute,
	}
	if err := srv.Serve(listen); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
