package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type ExecOptions struct {
	Scripts map[string]CmdOptions `json:"scripts"`
}

type CmdOptions struct {
	Command []string
	Timeout time.Duration
}

func (c *CmdOptions) UnmarshalJSON(data []byte) error {
	var opt struct {
		Command []string `json:"command"`
		Timeout string   `json:"timeout"`
	}
	if err := json.Unmarshal(data, &opt); err != nil {
		return err
	}
	d, err := time.ParseDuration(opt.Timeout)
	if err != nil {
		return errors.Wrap(err, "time.ParseDuration")
	}
	c.Command = opt.Command
	c.Timeout = d
	return nil
}

type ExecService struct {
	opts   ExecOptions
	active map[string]bool
	mu     sync.Mutex
}

func NewExecService(opts ExecOptions) *ExecService {
	return &ExecService{
		opts:   opts,
		active: make(map[string]bool, len(opts.Scripts)),
	}
}

func (s *ExecService) handleExec(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/exec/")
	script, ok := s.opts.Scripts[id]
	if !ok {
		http.Error(w, "please specify a valid id", http.StatusBadRequest)
		return
	}
	fmt.Fprintln(w, "ok")

	go func() {
		if s.setActive(id) {
			log.Printf("script %q is already running", id)
			return
		}
		defer s.unsetActive(id)

		ctx, cancel := context.WithTimeout(context.Background(), script.Timeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, script.Command[0], script.Command[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Printf("failed to runs script %q: %v", id, err)
		}
	}()
}

func (s *ExecService) setActive(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active[id] {
		return false
	}
	s.active[id] = true
	return true
}

func (s *ExecService) unsetActive(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.active[id] = false
}
