package k8sresolver

import (
	"context"
	"time"
)

func debounce(ctx context.Context, dur time.Duration, in chan interface{}, cb func()) {
	// create a stopped timer
	t := time.NewTimer(0)
	if !t.Stop() {
		<-t.C
	}
	// the debounce loop
	for {
		select {
		case _ = <-in:
			t.Reset(dur)
		case <-t.C:
			cb()
		case <-ctx.Done():
			return
		}
	}
}

func strSliceEqual(strs1 []string, strs2 []string) bool {
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
