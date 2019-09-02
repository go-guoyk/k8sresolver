package k8sresolver

import (
	"context"
	"time"
)

func debounce(ctx context.Context, dur time.Duration, in chan interface{}, cb func()) {
	update := false
	timer := time.NewTimer(dur)
	for {
		select {
		case _ = <-in:
			update = true
			timer.Reset(dur)
		case <-timer.C:
			if update {
				cb()
			}
		case <-ctx.Done():
			return
		}
	}
}

func StrSliceEqual(strs1 []string, strs2 []string) bool {
	if len(strs1) != len(strs2) {
		return false
	}
	for i := 0; i < len(strs1); i++ {
		if strs1[i] != strs2[i] {
			return false
		}
	}
	return true
}
