package mcp

import (
	"context"
	"errors"

	"github.com/Mayank2930/bruno-mcp-server/internal/workspace"
)

type InitializeParams struct {
	ClientInfo map[string]any `json:"clientInfo,omitempty"`
}

type ToolCallParams struct {
	Name      string         `json:"name"`
	Tool      string         `json:"tool:omitempty"`
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
				"name":        "workspace.get",
				"description": "Get a registered workspace by name",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
					},
					"required": []string{"name"}, // <-- FIXED
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
				"name":        "workspace.list",
				"description": "List registered workspaces",
				"inputSchema": map[string]any{
					"type":       "object",
					"properties": map[string]any{}, // no args
				},
				"outputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"workspaces": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"name": map[string]any{"type": "string"},
									"path": map[string]any{"type": "string"},
								},
							},
						},
					},
				},
			},
			{
				"name":        "collections.list",
				"description": "List collections in a workspace",
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
			{
				"name":        "requests.list",
				"description": "List requests in a collection",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"workspace":  map[string]any{"type": "string"},
						"collection": map[string]any{"type": "string"},
					},
					"required": []string{"workspace", "collection"},
				},
				"outputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"requests": map[string]any{
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

	toolName := params.Name
	if toolName == "" {
		toolName = params.Tool
	}
	if toolName == "" {
		return nil, NewError(CodeInvalidParams, "Invalid params: name is required")
	}

	switch toolName {

	case "workspace.register":
		nameAny, ok := params.Arguments["name"]
		if !ok {
			return nil, NewError(CodeInvalidParams, "Invalid params: name is required")
		}
		pathAny, ok := params.Arguments["path"]
		if !ok {
			return nil, NewError(CodeInvalidParams, "Invalid params: path is required")
		}

		name, _ := nameAny.(string)
		path, _ := pathAny.(string)

		createIfMissing := false
		if v, ok := params.Arguments["createIfMissing"]; ok {
			if b, ok := v.(bool); ok {
				createIfMissing = b
			}
		}

		ws, err := s.registry.Register(name, path, createIfMissing)
		if err != nil {
			return nil, workspaceToRPCError(err)
		}
		return ws, nil

	case "workspace.get":
		nameAny, ok := params.Arguments["name"]
		if !ok {
			return nil, NewError(CodeInvalidParams, "Invalid params: name is required")
		}
		name, _ := nameAny.(string)

		ws, err := s.registry.Get(name)
		if err != nil {
			return nil, workspaceToRPCError(err)
		}
		return ws, nil

	case "workspace.list":
		return map[string]any{
			"workspaces": s.registry.List(),
		}, nil

	case "collections.list":
		wsNameAny, ok := params.Arguments["workspace"]
		if !ok {
			return nil, NewError(CodeInvalidParams, "Invalid params: workspace is required")
		}
		wsName, _ := wsNameAny.(string)

		ws, err := s.registry.Get(wsName)
		if err != nil {
			return nil, workspaceToRPCError(err)
		}

		cols, err := s.bruno.ListCollections(ws.Path)
		if err != nil {
			return nil, NewError(CodeInternalError, err.Error())
		}

		return map[string]any{"collections": cols}, nil

	case "requests.list":
		wsNameAny, ok := params.Arguments["workspace"]
		if !ok {
			return nil, NewError(CodeInvalidParams, "Invalid params: workspace is required")
		}
		colAny, ok := params.Arguments["collection"]
		if !ok {
			return nil, NewError(CodeInvalidParams, "Invalid params: collection is required")
		}
		wsName, _ := wsNameAny.(string)
		collection, _ := colAny.(string)

		ws, err := s.registry.Get(wsName)
		if err != nil {
			return nil, workspaceToRPCError(err)
		}

		reqs, err := s.bruno.ListRequests(ws.Path, collection)
		if err != nil {
			return nil, NewError(CodeInternalError, err.Error())
		}

		return map[string]any{"requests": reqs}, nil

	default:
		return nil, NewError(CodeMethodNotFound, "Unknown tool: "+toolName)
	}
}

func workspaceToRPCError(err error) *RPCError {
	switch {
	case errors.Is(err, workspace.ErrInvalidName),
		errors.Is(err, workspace.ErrInvalidPath):
		return NewError(CodeInvalidParams, err.Error())
	case errors.Is(err, workspace.ErrNotFound):
		return NewError(CodeInvalidParams, err.Error())
	default:
		return NewError(CodeInternalError, err.Error())
	}
}
