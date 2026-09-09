package sandbox

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const selfSignalBaseURL = "http://127.0.0.1:1029"

// SelfPause asks the local sandbox agent to pause its own sandbox.
func SelfPause(ctx context.Context, reason string) error {
	return selfSignal(ctx, "pause", reason)
}

// SelfDelete asks the local sandbox agent to irreversibly delete its own sandbox.
func SelfDelete(ctx context.Context, reason string) error {
	return selfSignal(ctx, "delete", reason)
}

func selfSignal(ctx context.Context, action, reason string) error {
	endpoint, err := url.Parse(selfSignalBaseURL + "/self/" + action)
	if err != nil {
		return fmt.Errorf("build self-%s URL: %w", action, err)
	}
	if reason != "" {
		query := endpoint.Query()
		query.Set("reason", reason)
		endpoint.RawQuery = query.Encode()
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("create self-%s request: %w", action, err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("send self-%s request: %w", action, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusAccepted {
		return fmt.Errorf("self-%s request returned HTTP %d", action, response.StatusCode)
	}
	return nil
}
