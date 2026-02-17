package mcp

import "encoding/json"

const VERSION = "2.0"

type ID = json.RawMessage

type Request struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      ID               `json:"id,omitempty"`
	Method  string           `json:"method"`
	Params  *json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      ID        `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

type RPCError struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data,omitempty"`
}

func (r Request) IsNotification() bool {
	if len(r.ID) == 0 {
		return true
	}
	return string(r.ID) == "null"
}
