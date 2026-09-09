package protocol

import (
	"errors"
	"io"
	"strings"
	"testing"
)

type closeTracker struct {
	io.Reader
	closed bool
}

func (c *closeTracker) Close() error {
	c.closed = true
	return nil
}

func TestNDJSONDecoder(t *testing.T) {
	t.Parallel()
	body := &closeTracker{Reader: strings.NewReader(": ping\nevent: output\ndata: {\"value\":1}\n{\"value\":2}\n")}
	decoder := NewNDJSONDecoder[struct {
		Value int `json:"value"`
	}](body)

	first, err := decoder.Recv()
	if err != nil || first.Value != 1 {
		t.Fatalf("first Recv() = %#v, %v", first, err)
	}
	second, err := decoder.Recv()
	if err != nil || second.Value != 2 {
		t.Fatalf("second Recv() = %#v, %v", second, err)
	}
	if _, err := decoder.Recv(); !errors.Is(err, io.EOF) {
		t.Fatalf("final Recv() error = %v, want io.EOF", err)
	}
	if !body.closed {
		t.Fatal("body was not closed after EOF")
	}
}

func TestNDJSONDecoderRejectsInvalidJSON(t *testing.T) {
	t.Parallel()
	decoder := NewNDJSONDecoder[map[string]any](io.NopCloser(strings.NewReader("not-json\n")))
	if _, err := decoder.Recv(); err == nil {
		t.Fatal("Recv() error = nil, want decoding error")
	}
}
