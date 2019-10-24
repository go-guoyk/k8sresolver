package k8sresolver

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDebounce(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	requests := make(chan interface{}, 1)

	var val int
	go debounce(ctx, time.Second, requests, func() { val++ })

	requests <- nil
	requests <- nil
	time.Sleep(time.Millisecond * 900)
	assert.Equal(t, 1, val, "should not increase")

	requests <- nil
	time.Sleep(time.Millisecond * 900)
	assert.Equal(t, 1, val, "should not increase")

	requests <- nil
	time.Sleep(time.Millisecond * 900)
	assert.Equal(t, 1, val, "should not increase")

	requests <- nil
	time.Sleep(time.Millisecond * 1100)
	assert.Equal(t, 2, val, "should increase")

	time.Sleep(time.Millisecond * 1100)
	assert.Equal(t, 2, val, "should not increase again")
}

func TestSliceStrEqual(t *testing.T) {
	assert.True(t, strSliceEqual(nil, []string{}), "case 1")
	assert.False(t, strSliceEqual(nil, []string{"1"}), "case 2")
	assert.True(t, strSliceEqual([]string{"1"}, []string{"1"}), "case 3")
}
