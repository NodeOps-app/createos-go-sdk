package protocol

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
)

const maxNDJSONTokenSize = 16 << 20

// NDJSONDecoder incrementally decodes newline-delimited JSON. It also accepts
// SSE data lines and ignores SSE control and comment lines.
type NDJSONDecoder[T any] struct {
	scanner *bufio.Scanner
	body    io.ReadCloser
	once    sync.Once
}

// NewNDJSONDecoder returns a decoder that owns body. Call Close when stopping
// before io.EOF so the underlying HTTP connection is released.
func NewNDJSONDecoder[T any](body io.ReadCloser) *NDJSONDecoder[T] {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64*1024), maxNDJSONTokenSize)
	return &NDJSONDecoder[T]{scanner: scanner, body: body}
}

// Recv returns the next event. It returns io.EOF after a cleanly drained body.
func (d *NDJSONDecoder[T]) Recv() (T, error) {
	var zero T
	for d.scanner.Scan() {
		line := strings.TrimSpace(d.scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") ||
			strings.HasPrefix(line, "event:") || strings.HasPrefix(line, "id:") ||
			strings.HasPrefix(line, "retry:") {
			continue
		}
		if strings.HasPrefix(line, "data:") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if line == "" {
				continue
			}
		}

		var event T
		if err := json.NewDecoder(bytes.NewBufferString(line)).Decode(&event); err != nil {
			_ = d.Close()
			return zero, fmt.Errorf("decode NDJSON event: %w", err)
		}
		return event, nil
	}
	if err := d.scanner.Err(); err != nil {
		_ = d.Close()
		return zero, fmt.Errorf("read NDJSON stream: %w", err)
	}
	_ = d.Close()
	return zero, io.EOF
}

// Close closes the underlying response body. It is safe to call repeatedly.
func (d *NDJSONDecoder[T]) Close() error {
	var err error
	d.once.Do(func() { err = d.body.Close() })
	return err
}
