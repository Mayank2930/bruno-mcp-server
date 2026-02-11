package mcp

import "context"

type InitializeParams struct {
	ClientInfo map[string]any `json:"clientInfo,omitempty"`
}

type ToolCallParams struct {
	Tool      string         `json:"tool"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

func (s *Server) RegisterCoreMethods() {
	s.Handle("initialize", s.handleInitialize)
	s.Handle("tools/list", s.handleToolList)
	s.Handle("tools/call", s.handleToolsCall)
}

func (s *Server) handleInitialize(ctx context.Context, req Request) (any, *RPCError) {
	_, _ = decodeParamsOptional[InitializeParams](req)

	return map[string]any{
		"serverInfo": map[string]any{
			"name":    "mcp-server-bruno",
			"version": "0.1.0",
		},
		"capabilities": map[string]any{
			"tools": map[string]any{
				"list": true,
				"call": true,
			},
		},
	}, nil
}

func (s *Server) handleToolList(ctx context.Context, req Request) (any, *RPCError) {
	return map[string]any{
		"tools": []map[string]any{
			{
				"name":        "workspace.register",
				"description": "Register a workspace by name (stub in Step 2)",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name":            map[string]any{"type": "string"},
						"path":            map[string]any{"type": "string"},
						"createIfMissing": map[string]any{"type": "boolean"},
					},
					"required": []string{"name", "path"},
				},
				"outputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
						"path": map[string]any{"type": "string"},
					},
				},
			},
			{
				"name":        "collections.list",
				"description": "List collections in a workspace (stub in Step 2)",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"workspace": map[string]any{"type": "string"},
					},
					"required": []string{"workspace"},
				},
				"outputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"collections": map[string]any{
							"type":  "array",
							"items": map[string]any{"type": "string"},
						},
					},
				},
			},
		},
	}, nil
}

func (s *Server) handleToolsCall(ctx context.Context, req Request) (any, *RPCError) {
	params, rpcErr := decodeParams[ToolCallParams](req)
	if rpcErr != nil {
		return nil, rpcErr
	}
	if params.Tool == "" {
		return nil, NewError(CodeInvalidParams, "Invalid params: tool is required")
	}

	switch params.Tool {
	case "workspace.register":
		return map[string]any{
			"tool": params.Tool,
			"output": map[string]any{
				"status": "stub",
				"note":   "workspace.register will be implemented in Step 3",
			},
		}, nil
	case "collections.list":
		return map[string]any{
			"tool": params.Tool,
			"output": map[string]any{
				"status":      "stub",
				"collections": []string{},
				"note":        "collections.list will be implemented after Bruno adapter is added",
			},
		}, nil
	default:
		return nil, NewError(CodeMethodNotFound, "Unknown tool: "+params.Tool)
	}
}
