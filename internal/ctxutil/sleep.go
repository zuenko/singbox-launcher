// Package ctxutil provides context-aware utilities (e.g. sleep with cancellation).
package ctxutil

import (
	"context"
	"time"
)

// SleepWithContext blocks for duration d or until ctx is cancelled.
// Returns ctx.Err() if the context is cancelled before or during the wait (e.g. context.Canceled).
// Returns nil if d elapses without cancellation.
// Uses a timer internally so it is safe to call in loops without leaking timers.
func SleepWithContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer func() {
		if !t.Stop() {
			select {
			case <-t.C:
			default:
			}
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
