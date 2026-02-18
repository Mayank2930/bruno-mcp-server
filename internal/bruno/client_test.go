package bruno

import (
	"context"
	"testing"
)

func TestRunErrorsWhenBruMissing(t *testing.T) {
	c := &Client{brunoPath: ""}
	_, stderr, err := c.run(context.Background(), "--version")
	if err == nil {
		t.Fatalf("expected error when bruPath is empty")
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr when bru is missing, got %q", stderr)
	}
}

func TestRunCapturesStderrOnFailure(t *testing.T) {
	c := NewClient()
	if !c.hasCli() {
		t.Skip("bru not found in PATH; install bruno or add it to PATH to run this test")
	}

	_, stderr, err := c.run(context.Background(), "definitely-not-a-real-command")
	if err == nil {
		t.Fatalf("expected error for invalid bru command")
	}
	if stderr == "" {
		t.Fatalf("expected stderr to be captured, got empty")
	}
}
