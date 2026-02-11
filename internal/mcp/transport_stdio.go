package mcp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
)

const (
	initialScanBuf = 1024 * 1024
	maxScanBuf     = 32 * 1024 * 1024
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

	if err := in.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			fmt.Println(s.stderr, "input line is too long: increase scanner size or send smaller requests")
		}
		return err
	}

	return nil
}
