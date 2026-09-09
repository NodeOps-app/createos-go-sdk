package sandbox

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// Egress returns the current egress allowlist.
func (i *Instance) Egress(ctx context.Context) (structs.EgressView, error) {
	var response structs.EgressView
	if err := i.transport.Do(ctx, http.MethodGet, i.path("/egress"), transport.RequestOptions{}, &response); err != nil {
		return structs.EgressView{}, fmt.Errorf("get egress for sandbox %q: %w", i.ID(), err)
	}
	return response, nil
}

// SetEgress replaces the egress allowlist. A nil or empty slice allows all egress.
func (i *Instance) SetEgress(ctx context.Context, rules []string) (structs.EgressView, error) {
	var response structs.EgressView
	request := structs.SetEgressRequest{EgressRules: &rules}
	err := i.transport.Do(ctx, http.MethodPut, i.path("/egress"), transport.RequestOptions{Body: request}, &response)
	if err != nil {
		return structs.EgressView{}, fmt.Errorf("set egress for sandbox %q: %w", i.ID(), err)
	}
	return response, nil
}
