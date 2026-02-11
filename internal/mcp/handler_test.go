package mcp

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDispatch_MethodNotFound(t *testing.T) {
	s := NewServer()
	req := Request{JSONRPC: VERSION, ID: json.RawMessage("1"), Method: "nope"}

	_, rpcErr := s.dispatch(context.Background(), req)
	if rpcErr == nil {
		t.Fatalf("expected error")
	}
	if rpcErr.Code != CodeMethodNotFound {
		t.Fatalf("expected CodeMethodNotFound got %d", rpcErr.Code)
	}
}

func TestInitialize_Handler(t *testing.T) {
	s := NewServer()
	s.RegisterCoreMethods()

	req := Request{JSONRPC: VERSION, ID: json.RawMessage("1"), Method: "initialize"}
	res, rpcErr := s.dispatch(context.Background(), req)
	if rpcErr != nil {
		t.Fatalf("expected nil error got %+v", rpcErr)
	}

	m, ok := res.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", res)
	}
	if m["serverInfo"] == nil {
		t.Fatalf("expected serverInfo in result")
	}
	if m["capabilities"] == nil {
		t.Fatalf("expected capabilities in result")
	}
}

func TestToolsList_Handler(t *testing.T) {
	s := NewServer()
	s.RegisterCoreMethods()

	req := Request{JSONRPC: VERSION, ID: json.RawMessage("2"), Method: "tools/list"}
	res, rpcErr := s.dispatch(context.Background(), req)
	if rpcErr != nil {
		t.Fatalf("expected nil error got %+v", rpcErr)
	}

	m, ok := res.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", res)
	}

	toolsVal, ok := m["tools"]
	if !ok {
		t.Fatalf("expected tools key in result")
	}

	tools, ok := toolsVal.([]any)
	if !ok {
		t.Fatalf("expected tools to be an array, got %T", toolsVal)
	}
	if len(tools) == 0 {
		t.Fatalf("expected at least one tool")
	}
}

func TestToolsCall_MissingParams(t *testing.T) {
	s := NewServer()
	s.RegisterCoreMethods()

	req := Request{JSONRPC: VERSION, ID: json.RawMessage("3"), Method: "tools/call"} // no params
	_, rpcErr := s.dispatch(context.Background(), req)
	if rpcErr == nil {
		t.Fatalf("expected invalid params error")
	}
	if rpcErr.Code != CodeInvalidParams {
		t.Fatalf("expected CodeInvalidParams got %d", rpcErr.Code)
	}
}

func TestToolsCall_UnknownTool(t *testing.T) {
	s := NewServer()
	s.RegisterCoreMethods()

	p := json.RawMessage(`{"tool":"nope","arguments":{}}`)
	req := Request{JSONRPC: VERSION, ID: json.RawMessage("4"), Method: "tools/call", Params: &p}

	_, rpcErr := s.dispatch(context.Background(), req)
	if rpcErr == nil {
		t.Fatalf("expected method not found error for unknown tool")
	}
	if rpcErr.Code != CodeMethodNotFound {
		t.Fatalf("expected CodeMethodNotFound got %d", rpcErr.Code)
	}
}
