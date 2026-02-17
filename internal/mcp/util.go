package mcp

import (
	"bufio"
	"encoding/json"
)

func writeResponse(w *bufio.Writer, resp Response) error {
	b, err := json.Marshal(resp)

	if err != nil {
		return err
	}

	if _, err := w.Write(b); err != nil {
		return err
	}

	return w.WriteByte('\n')
}

func parseAndValidateRequest(line []byte) (Request, *RPCError) {
	var base struct {
		JSONRPC json.RawMessage `json:"jsonrpc"`
		ID      json.RawMessage `json:"id,omitempty"`
		Method  json.RawMessage `json:"method"`
		Params  json.RawMessage `json:"params,omitempty"`
	}

	if err := json.Unmarshal(line, &base); err != nil {
		return Request{}, NewError(CodeParseError, "Parse Error!")
	}

	req := Request{ID: ID(base.ID)}

	var ver string
	if len(base.JSONRPC) == 0 || json.Unmarshal(base.JSONRPC, &ver) != nil || ver != VERSION {
		return req, NewError(CodeInvalidRequest, "Invalid Request: jsonRpc must be \"2.0\"")
	}

	req.JSONRPC = ver

	if len(base.Method) == 0 {
		return req, NewError(CodeInvalidRequest, "Invalid method or missing method")
	}

	var method string
	if err := json.Unmarshal(base.Method, &method); err != nil || method == "" {
		return req, NewError(CodeInvalidRequest, "Invalid Request: method must be a non-empty string")
	}

	req.Method = method

	if len(base.Params) != 0 {
		p := json.RawMessage(append([]byte(nil), base.Params...))
		req.Params = &p
	}

	return req, nil

}
