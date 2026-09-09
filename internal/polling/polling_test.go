package polling

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWait(t *testing.T) {
	t.Parallel()
	calls := 0
	err := Wait(context.Background(), Options{Interval: time.Nanosecond}, func(context.Context) (bool, error) {
		calls++
		return calls == 3, nil
	})
	if err != nil || calls != 3 {
		t.Fatalf("Wait() = %v after %d calls", err, calls)
	}
}

func TestWaitTimeout(t *testing.T) {
	t.Parallel()
	err := Wait(context.Background(), Options{Interval: time.Millisecond, Timeout: time.Millisecond}, func(context.Context) (bool, error) {
		return false, nil
	})
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("Wait() = %v, want ErrTimeout", err)
	}
}
