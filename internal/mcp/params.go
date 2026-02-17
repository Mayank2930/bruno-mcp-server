package mcp

import "encoding/json"

func decodeParams[T any](req Request) (T, *RPCError) {
	var zero T
	if req.Params == nil {
		return zero, NewError(CodeInvalidParams, "Invalid Params: Missing params")
	}

	if err := json.Unmarshal(*req.Params, &zero); err != nil {
		return zero, NewError(CodeInvalidParams, "Invalid Params: Malformed params")
	}

	return zero, nil
}

func decodeParamsOptional[T any](req Request) (T, *RPCError) {
	var zero T
	if req.Params == nil {
		return zero, nil
	}

	if err := json.Unmarshal(*req.Params, &zero); err != nil {
		return zero, NewError(CodeInvalidParams, "Invalid Params: Malformed params")
	}

	return zero, nil
}
