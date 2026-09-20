//go:build !(darwin || dragonfly || freebsd || linux || netbsd || openbsd || windows)

package definitionupdate

import (
	"context"
	"fmt"
	"runtime"
)

func acquireLock(ctx context.Context, _ string) (func() error, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("managed definition cache locking is unsupported on %s", runtime.GOOS)
}
