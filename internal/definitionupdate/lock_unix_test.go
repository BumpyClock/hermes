//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package definitionupdate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAcquireLockCancellationAndOwnerReleaseRecovery(t *testing.T) {
	root, err := os.MkdirTemp(".", ".definition-lock-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(root) })
	path := filepath.Join(root, "cache.lock")

	unlock, err := acquireLock(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	waiting, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if _, err = acquireLock(waiting, path); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("lock wait error = %v, want deadline exceeded", err)
	}
	if err = unlock(); err != nil {
		t.Fatal(err)
	}
	recovered, err := acquireLock(context.Background(), path)
	if err != nil {
		t.Fatalf("lock did not recover after owner release: %v", err)
	}
	if err = recovered(); err != nil {
		t.Fatal(err)
	}
}
