package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Mayank2930/bruno-mcp-server/internal/mcp"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	s := mcp.NewServer()
	s.RegisterCoreMethods()

	if err := s.ServeStdio(ctx); err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
