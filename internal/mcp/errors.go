package mcp

const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

func NewError(code int, message string) *RPCError {
	return &RPCError{Code: code, Message: message}
}

func NewErrorWithData(code int, message string, data map[string]any) *RPCError {
	return &RPCError{Code: code, Message: message, Data: data}
}
