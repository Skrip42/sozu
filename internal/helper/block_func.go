package helper

import "context"

func NewSyncer(ctx context.Context) (func(), func()) {
	ch := make(chan struct{})
	done := func() {
		select {
		case ch <- struct{}{}:
		case <-ctx.Done():
		}
	}
	await := func() {
		select {
		case <-ch:
		case <-ctx.Done():
		}
	}
	return await, done
}
