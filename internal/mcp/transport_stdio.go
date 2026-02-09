package mcp

import (
	"bufio"
	"context"
	"os"
)

const (
	initialScanBuf = 1024 * 1024
	maxScanBuf     = 23 * 1024 * 1024
)

func (s *Server) ServeStdio(ctx context.Context) error {
	in := bufio.NewScanner(os.Stdin)

	buf := make([]byte, initialScanBuf)
	in.Buffer(buf, maxScanBuf)

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	for in.Scan() {
		line := in.Bytes()
		if len(line) == 0 {
			continue
		}

		req, rpcError := parseAndValidateRequest(line)
		if rpcError != nil {
			if !req.IsNotification() {
				_ = writeResponse(out, Response{
					JSONRPC: VERSION,
					ID:      req.ID,
					Error:   rpcError,
				})
				_ = out.Flush()
			}
			continue
		}

		result, errObj := s.dispatch(ctx, req)
		if req.IsNotification() {
			continue
		}

		resp := Response{
			JSONRPC: VERSION,
			ID:      req.ID,
		}

		if errObj != nil {
			resp.Error = errObj
		} else {
			resp.Result = result
		}

		if err := writeResponse(out, resp); err != nil {
			return err
		}

		_ = out.Flush()
	}

	return in.Err()
}
