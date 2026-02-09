package mcp

import "encoding/json"

const VERSION = "2.0"

type ID = json.RawMessage

type Request struct {
	JSONRPC string
	ID      ID
	Method  string
	Params  *json.RawMessage
}

type Response struct {
	JSONRPC string
	ID      ID
	Result  any
	Error   *RPCError
}

type RPCError struct {
	Code    int
	Message string
	Data    map[string]any
}

func (r Request) IsNotification() bool {
	if len(r.ID) == 0 {
		return true
	}

	return string(r.ID) == "null"
}
