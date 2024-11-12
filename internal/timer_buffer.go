package internal

import (
	"context"
	"time"
)

type timerFactory[V any] struct {
	base      Factory[V]
	duratiorn time.Duration
}

func NewTimerFactory[V any](base Factory[V], duration time.Duration) Factory[V] {
	return &timerFactory[V]{base: base, duratiorn: duration}
}

func (f *timerFactory[V]) Create(
	ctx context.Context,
	inputCh chan V,
	flushCh chan func(),
	beforeSend func(V, func()),
	afterSend func(V, func()),
	beforeFlush func(),
	afterFlush func(int),
	capacity int,
	cancel context.CancelFunc,
) <-chan []V {
	counter := 0
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
		counter++
		afterSend(item, flush)
	}
	flush := func(flushCount int) {
		counter -= flushCount
		if counter == 0 {
			stopTimer()
		}
		afterFlush(flushCount)
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

	return f.base.Create(
		ctx,
		inputCh,
		flushCh,
		beforeSend,
		send,
		beforeFlush,
		flush,
		capacity,
		cancel,
	)
}
