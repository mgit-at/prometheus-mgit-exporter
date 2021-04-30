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
	} `json:"fstab"`
	BinLog struct {
		Enable bool `json:"enable"`
		MySQLBinOptions
	} `json:"binlog"`
	RasDaemon struct {
		Enable bool `json:"enable"`
		RasDaemonOptions
	} `json:"rasdaemon"`
	Elk struct {
		Enable bool `json:"enable"`
		ElkOptions
	} `json:"elk"`
	Exec struct {
		Enable bool `json:"enable"`
		ExecOptions
	} `json:"exec"`
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

	if cfg.BinLog.Enable {
		log.Println("enabling binlog checker")
		c := NewMySQLBinChecker(cfg.BinLog.MySQLBinOptions)
		if err := prometheus.Register(c); err != nil {
			return fmt.Errorf("failed to register binlog checker: %v", err)
		}
	}

	if cfg.RasDaemon.Enable {
		log.Println("enabling rasdaemon checker")
		c := NewRasdaemonChecker(cfg.RasDaemon.RasDaemonOptions)
		if err := prometheus.Register(c); err != nil {
			return fmt.Errorf("failed to register rasdaemon checker: %v", err)
		}
	}

	if cfg.Elk.Enable {
		log.Println("enabling elk checker")
		c := NewElkChecker(cfg.Elk.ElkOptions)
		if err := prometheus.Register(c); err != nil {
			return fmt.Errorf("failed to register elk checker: %v", err)
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

	if cfg.Exec.Enable {
		log.Println("enabling automatic execution of services on prometheus alert")
		e := NewExecService(cfg.Exec.ExecOptions)
		http.Handle("/exec/", http.HandlerFunc(e.handleExec))
	}

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
