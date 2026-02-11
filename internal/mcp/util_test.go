package mcp

import (
	"encoding/json"
	"testing"
)

func TestParseAndValidateRequest_Valid(t *testing.T) {
	line := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"clientInfo":{"name":"x"}}}`)

	req, rpcErr := parseAndValidateRequest(line)
	if rpcErr != nil {
		t.Fatalf("expected nil error, got %+v", rpcErr)
	}
	if req.JSONRPC != VERSION {
		t.Fatalf("expected jsonrpc=%s got %s", VERSION, req.JSONRPC)
	}
	if req.Method != "initialize" {
		t.Fatalf("expected method initialize got %s", req.Method)
	}
	if string(req.ID) != "1" {
		t.Fatalf("expected id raw '1', got %s", string(req.ID))
	}
	if req.Params == nil {
		t.Fatalf("expected params non-nil")
	}
}

func TestParseAndValidateRequest_InvalidJSON(t *testing.T) {
	line := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"`) // missing brace

	_, rpcErr := parseAndValidateRequest(line)
	if rpcErr == nil {
		t.Fatalf("expected parse error, got nil")
	}
	if rpcErr.Code != CodeParseError {
		t.Fatalf("expected CodeParseError, got %d", rpcErr.Code)
	}
}

func TestParseAndValidateRequest_InvalidVersion(t *testing.T) {
	line := []byte(`{"jsonrpc":"1.0","id":1,"method":"initialize"}`)

	_, rpcErr := parseAndValidateRequest(line)
	if rpcErr == nil {
		t.Fatalf("expected invalid request error")
	}
	if rpcErr.Code != CodeInvalidRequest {
		t.Fatalf("expected CodeInvalidRequest got %d", rpcErr.Code)
	}
}

func TestParseAndValidateRequest_MissingMethod(t *testing.T) {
	line := []byte(`{"jsonrpc":"2.0","id":1}`)

	_, rpcErr := parseAndValidateRequest(line)
	if rpcErr == nil {
		t.Fatalf("expected invalid request error")
	}
	if rpcErr.Code != CodeInvalidRequest {
		t.Fatalf("expected CodeInvalidRequest got %d", rpcErr.Code)
	}
}

func TestRequest_IsNotification(t *testing.T) {
	// missing id
	req1 := Request{JSONRPC: VERSION, Method: "initialize"}
	if !req1.IsNotification() {
		t.Fatalf("expected notification when id missing")
	}

	// id null
	req2 := Request{JSONRPC: VERSION, Method: "initialize", ID: json.RawMessage("null")}
	if !req2.IsNotification() {
		t.Fatalf("expected notification when id is null")
	}

	// id present
	req3 := Request{JSONRPC: VERSION, Method: "initialize", ID: json.RawMessage("1")}
	if req3.IsNotification() {
		t.Fatalf("expected non-notification when id present")
	}
}
