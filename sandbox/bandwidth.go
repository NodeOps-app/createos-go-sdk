package sandbox

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// Bandwidth returns the sandbox bandwidth quota and usage.
func (i *Instance) Bandwidth(ctx context.Context) (structs.BandwidthView, error) {
	var response structs.BandwidthView
	if err := i.transport.Do(ctx, http.MethodGet, i.path("/bandwidth"), transport.RequestOptions{}, &response); err != nil {
		return structs.BandwidthView{}, fmt.Errorf("get bandwidth for sandbox %q: %w", i.ID(), err)
	}
	return response, nil
}

// RechargeBandwidth adds bytes to the sandbox quota.
func (i *Instance) RechargeBandwidth(ctx context.Context, bytes int64) (structs.BandwidthView, error) {
	var response structs.BandwidthView
	request := structs.RechargeBandwidthRequest{AddBytes: bytes}
	err := i.transport.Do(ctx, http.MethodPost, i.path("/bandwidth/recharge"), transport.RequestOptions{Body: request}, &response)
	if err != nil {
		return structs.BandwidthView{}, fmt.Errorf("recharge bandwidth for sandbox %q: %w", i.ID(), err)
	}
	return response, nil
}
