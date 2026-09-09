// Package protocol implements the control-plane wire protocols.
package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// ErrEmptyEnvelope indicates that an endpoint expected to return JSend data
// returned no response body.
var ErrEmptyEnvelope = errors.New("empty JSend envelope")

// EnvelopeError describes a valid JSend envelope whose status is not success.
type EnvelopeError struct {
	Status  string
	Message string
	Code    int
	Data    json.RawMessage
}

func (e *EnvelopeError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("unexpected JSend status %q", e.Status)
}

// UnmarshalJSend unwraps and decodes a JSend success envelope using the
// endpoint-specific response type T.
func UnmarshalJSend[T any](body []byte) (T, error) {
	var zero T
	if len(bytes.TrimSpace(body)) == 0 {
		return zero, ErrEmptyEnvelope
	}

	var env structs.JSendEnvelope[json.RawMessage]
	if err := json.Unmarshal(body, &env); err != nil {
		return zero, fmt.Errorf("decode JSend envelope: %w", err)
	}
	if env.Status != "success" {
		message := env.Message
		if len(env.Data) != 0 && string(env.Data) != "null" {
			var details struct {
				Message string `json:"message"`
			}
			if json.Unmarshal(env.Data, &details) == nil {
				if message == "" {
					message = details.Message
				}
			}
		}
		return zero, &EnvelopeError{Status: env.Status, Message: message, Code: env.Code, Data: env.Data}
	}
	if len(env.Data) == 0 || string(env.Data) == "null" {
		return zero, nil
	}
	var result T
	if err := json.Unmarshal(env.Data, &result); err != nil {
		return zero, fmt.Errorf("decode JSend data: %w", err)
	}
	return result, nil
}
