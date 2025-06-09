package internal

import (
	"context"
	"sync"
)

type multiplexorFactory[V any, C comparable] struct {
	base      Factory[V]
	separator func(V) C
	capacity  int
}

func NewMultiplexorFactory[V any, C comparable](
	base Factory[V],
	separator func(V) C,
	capacity int,
) Factory[V] {
	return &multiplexorFactory[V, C]{
		base:      base,
		separator: separator,
		capacity:  capacity,
	}
}

func (f *multiplexorFactory[V, C]) Create(
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
	output := make(chan []V)

	flushMap := make(map[C]chan func(), f.capacity)
	inputMap := make(map[C]chan V, f.capacity)
	lastAfterFlushCount := make(map[C]int, f.capacity)

	outWg := sync.WaitGroup{}

	addBufer := func(selector C) chan V {
		inputMap[selector] = make(chan V)
		flushMap[selector] = make(chan func())
		lastAfterFlushCount[selector] = 0
		outputCh := f.base.Create(
			ctx,
			inputMap[selector],
			flushMap[selector],
			func(_ V, _ func()) {},
			func(_ V, _ func()) {},
			beforeFlush,
			func(count int) { lastAfterFlushCount[selector] = count },
			capacity,
			cancel,
		)
		outWg.Add(1)
		go func() {
			defer outWg.Done()
			for {
				items, ok := <-outputCh
				if !ok {
					return
				}
				output <- items
				afterFlush(lastAfterFlushCount[selector])
			}
		}()

		return inputMap[selector]
	}

	flush := func() {
		wg := sync.WaitGroup{}
		wg.Add(len(flushMap))
		for _, fl := range flushMap {
			go func() {
				select {
				case fl <- func() { wg.Done() }:
				case <-ctx.Done():
					wg.Done()
				}
			}()
		}
		wg.Wait()
	}

	go func() {
		defer func() {
			outWg.Wait()
			close(output)
		}()
		for {
			select {
			case item, ok := <-inputCh:
				if !ok {
					flush()
					for _, inch := range inputMap {
						close(inch)
					}
					return
				}
				selector := f.separator(item)
				input, ok := inputMap[selector]
				if !ok {
					input = addBufer(selector)
				}
				beforeSend(item, flush)
				select {
				case input <- item:
				case <-ctx.Done():
				}
				afterSend(item, flush)
			case done := <-flushCh:
				flush()
				done()
			case <-ctx.Done():
				flush()
				return
			}
		}
	}()

	return output
}
