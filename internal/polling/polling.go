// Package polling contains context-aware polling primitives.
package polling

import (
	"context"
	"errors"
	"time"
)

// ErrTimeout indicates that a poller's configured timeout elapsed.
var ErrTimeout = errors.New("polling timed out")

// Options configures Wait.
type Options struct {
	Interval time.Duration
	Timeout  time.Duration
}

// Sleep waits for duration or context cancellation.
func Sleep(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// Wait invokes check immediately and then at each interval until it succeeds,
// returns an error, the context is cancelled, or the timeout expires.
func Wait(ctx context.Context, opts Options, check func(context.Context) (bool, error)) error {
	if opts.Interval <= 0 {
		opts.Interval = 500 * time.Millisecond
	}
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeoutCause(ctx, opts.Timeout, ErrTimeout)
		defer cancel()
	}
	for {
		done, err := check(ctx)
		if err != nil || done {
			return err
		}
		if err := Sleep(ctx, opts.Interval); err != nil {
			return context.Cause(ctx)
		}
	}
}
