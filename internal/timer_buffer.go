package internal

import (
	"context"
	"time"
)

type timerFabric[V any] struct {
	base      Fabric[V]
	duratiorn time.Duration
}

func NewTimerFabric[V any](base Fabric[V], duration time.Duration) Fabric[V] {
	return &timerFabric[V]{base: base, duratiorn: duration}
}

func (f *timerFabric[V]) Create(
	ctx context.Context,
	inputCh chan V,
	flushCh chan func(),
	beforeSend func(V, func()),
	afterSend func(V, func()),
	beforeFlush func(),
	afterFlush func(),
	capacity int,
) <-chan []V {
	var tmStatus bool
	var tm *time.Timer
	stopTimer := func() {
		tmStatus = false
		ok := tm.Stop()
		if !ok {
			select {
			case <-tm.C:
			default:
			}
		}
	}
	startTimer := func() {
		tmStatus = true
		tm.Reset(f.duratiorn)
	}

	tm = time.NewTimer(f.duratiorn)
	stopTimer()

	send := func(item V, flush func()) {
		if !tmStatus {
			startTimer()
		}
		afterSend(item, flush)
	}
	flush := func() {
		stopTimer()
		afterFlush()
	}

	go func() {
		for {
			select {
			case <-tm.C:
				select {
				case flushCh <- func() {}:
				case <-ctx.Done():
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return f.base.Create(ctx, inputCh, flushCh, beforeSend, send, beforeFlush, flush, capacity)
}
