package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type ExecOptions struct {
	IDs map[string]CmdOptions `json:"ids"`
}

type CmdOptions struct {
	Timeout time.Duration `json:"duration"`
	Args    []string      `json:"args"`
}

type Service struct {
	opts   ExecOptions
	active map[string]bool
	mu     sync.Mutex
}

func NewExecService(opts ExecOptions) *Service {
	active := make(map[string]bool, len(opts.IDs))
	for id := range opts.IDs {
		active[id] = false
	}
	return &Service{
		opts:   opts,
		active: active,
	}
}

func (s *Service) handleExec(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/exec/")

	script, ok := s.opts.IDs[id]
	if !ok {
		http.Error(w, "please specify a valid id", http.StatusBadRequest)
		return
	}

	if s.setActive(id) {
		http.Error(w, fmt.Sprintf("script %s is already running", id), http.StatusConflict)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.opts.IDs[id].Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, script.Args[0], script.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Println("exec.Run", err)
		http.Error(w, fmt.Sprintf("failed to run script %s", id), http.StatusInternalServerError)
	}
	s.unsetActive(id)
}

func (s *Service) setActive(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active[id] {
		return false
	}
	s.active[id] = true
	return true
}

func (s *Service) unsetActive(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.active[id] = false
}
