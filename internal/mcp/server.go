package mcp

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/Mayank2930/bruno-mcp-server/internal/bruno"
	"github.com/Mayank2930/bruno-mcp-server/internal/workspace"
)

type HandlerFunc func(ctx context.Context, req Request) (any, *RPCError)

type Server struct {
	handlers map[string]HandlerFunc
	stderr   io.Writer
	registry *workspace.Registry
	bruno    *bruno.Client
}

func NewServer() *Server {
	return &Server{
		handlers: make(map[string]HandlerFunc),
		stderr:   os.Stderr,
		registry: workspace.NewRegistry(),
		bruno:    bruno.NewClient(),
	}
}

func (s *Server) Handle(method string, h HandlerFunc) {
	s.handlers[method] = h
}

func (s *Server) dispatch(ctx context.Context, req Request) (any, *RPCError) {
	h := s.handlers[req.Method]

	if h == nil {
		return nil, NewError(CodeMethodNotFound, "Method Not Found!")
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(s.stderr, "panic in the handler %q: %v \n", req.Method, r)
		}
	}()

	res, rpcError := h(ctx, req)
	if rpcError != nil {
		return nil, rpcError
	}
	return res, nil
}
