package mcp

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

var stdioMu sync.Mutex // because we temporarily replace os.Stdin/os.Stdout

func runServeStdio(t *testing.T, input string, s *Server) (string, error) {
	t.Helper()

	stdioMu.Lock()
	defer stdioMu.Unlock()

	// Save original stdio
	oldIn := os.Stdin
	oldOut := os.Stdout
	defer func() {
		os.Stdin = oldIn
		os.Stdout = oldOut
	}()

	// Create pipes
	inR, inW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe stdin: %v", err)
	}
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe stdout: %v", err)
	}

	os.Stdin = inR
	os.Stdout = outW

	// Run server
	done := make(chan error, 1)
	go func() {
		done <- s.ServeStdio(context.Background())
		_ = outW.Close() // allow reader to finish
	}()

	// Write input then close stdin writer
	_, _ = inW.WriteString(input)
	_ = inW.Close()

	// Read all stdout
	outBytes, _ := io.ReadAll(outR)
	_ = outR.Close()

	err = <-done
	return string(outBytes), err
}

func TestServeStdio_Initialize(t *testing.T) {
	s := NewServer()
	s.RegisterCoreMethods()

	out, err := runServeStdio(t, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`+"\n", s)
	if err != nil {
		t.Fatalf("ServeStdio returned error: %v", err)
	}
	out = strings.TrimSpace(out)
	if out == "" {
		t.Fatalf("expected output, got empty")
	}

	// Validate JSON-RPC response basics
	var resp Response
	if e := json.Unmarshal([]byte(out), &resp); e != nil {
		t.Fatalf("output is not valid JSON: %v\noutput=%s", e, out)
	}
	if resp.JSONRPC != VERSION {
		t.Fatalf("expected jsonrpc=%s got %s", VERSION, resp.JSONRPC)
	}
	if string(resp.ID) != "1" {
		t.Fatalf("expected id 1 got %s", string(resp.ID))
	}
	if resp.Error != nil {
		t.Fatalf("expected no error, got %+v", resp.Error)
	}
	if resp.Result == nil {
		t.Fatalf("expected result, got nil")
	}
}

func TestServeStdio_Notification_NoResponse(t *testing.T) {
	s := NewServer()
	s.RegisterCoreMethods()

	out, err := runServeStdio(t, `{"jsonrpc":"2.0","method":"initialize"}`+"\n", s)
	if err != nil {
		t.Fatalf("ServeStdio returned error: %v", err)
	}
	if strings.TrimSpace(out) != "" {
		t.Fatalf("expected no output for notification, got: %q", out)
	}
}
