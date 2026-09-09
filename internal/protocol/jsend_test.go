package protocol

import (
	"errors"
	"testing"
)

func TestUnmarshalJSend(t *testing.T) {
	t.Parallel()
	type response struct {
		ID string `json:"id"`
	}
	got, err := UnmarshalJSend[response]([]byte(`{"status":"success","data":{"id":"box-1"}}`))
	if err != nil {
		t.Fatalf("UnmarshalJSend() error = %v", err)
	}
	if got.ID != "box-1" {
		t.Fatalf("ID = %q, want box-1", got.ID)
	}
}

func TestUnmarshalJSendEnvelopeError(t *testing.T) {
	t.Parallel()
	_, err := UnmarshalJSend[struct{}]([]byte(`{"status":"error","message":"internal error","code":500}`))
	var envelopeErr *EnvelopeError
	if !errors.As(err, &envelopeErr) {
		t.Fatalf("error = %T, want *EnvelopeError", err)
	}
	if envelopeErr.Message != "internal error" || envelopeErr.Code != 500 {
		t.Fatalf("error = %#v", envelopeErr)
	}
}

func TestUnmarshalJSendEmpty(t *testing.T) {
	t.Parallel()
	_, err := UnmarshalJSend[struct{}](nil)
	if !errors.Is(err, ErrEmptyEnvelope) {
		t.Fatalf("error = %v, want ErrEmptyEnvelope", err)
	}
}
