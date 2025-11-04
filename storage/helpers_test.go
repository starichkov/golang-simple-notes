package storage

import (
	"context"
	"testing"
)

// CleanupClose registers t.Cleanup for anything with Close() error (no context).
func CleanupClose[T interface{ Close() error }](t testing.TB, c T) {
	t.Helper()
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Logf("cleanup: close failed: %v", err)
		}
	})
}

// CleanupCloseWithContext registers t.Cleanup for anything with Close(context.Context) error or Close() error.
// It logs errors instead of failing the test.
func CleanupCloseWithContext[T any](t testing.TB, ctx context.Context, c T) {
	t.Helper()
	t.Cleanup(func() {
		switch v := any(c).(type) {
		case interface{ Close(context.Context) error }:
			if err := v.Close(ctx); err != nil {
				t.Logf("cleanup: close failed: %v", err)
			}
		case interface{ Close() error }:
			if err := v.Close(); err != nil {
				t.Logf("cleanup: close failed: %v", err)
			}
		default:
			t.Logf("cleanup: unsupported closer type %T", c)
		}
	})
}
