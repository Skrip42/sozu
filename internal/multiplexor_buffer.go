package internal

import (
	"context"
	"sync"
)

type multiplexorFabric[V any, C comparable] struct {
	base      Fabric[V]
	separator func(V) C
	capacity  int
}

func NewMultiplezorFabric[V any, C comparable](
	base Fabric[V],
	separator func(V) C,
	capacity int,
) Fabric[V] {
	return &multiplexorFabric[V, C]{
		base:      base,
		separator: separator,
		capacity:  capacity,
	}
}

func (f *multiplexorFabric[V, C]) Create(
	ctx context.Context,
	inputCh chan V,
	flushCh chan func(),
	beforeSend func(V, func()),
	afterSend func(V, func()),
	beforeFlush func(),
	afterFlush func(int),
	capacity int,
) <-chan []V {
	runCtx, cancel := context.WithCancel(ctx)
	output := make(chan []V)

	flushMap := make(map[C]chan func(), f.capacity)
	inputMap := make(map[C]chan V, f.capacity)

	outWg := sync.WaitGroup{}

	addBufer := func(selector C) chan V {
		inputMap[selector] = make(chan V)
		flushMap[selector] = make(chan func())
		outputCh := f.base.Create(
			runCtx,
			inputMap[selector],
			flushMap[selector],
			func(_ V, _ func()) {},
			func(_ V, _ func()) {},
			beforeFlush,
			afterFlush,
			// func() {},
			// func(_ int) {},
			capacity,
		)
		outWg.Add(1)
		go func() {
			defer outWg.Done()
			for {
				select {
				case items, ok := <-outputCh:
					if !ok {
						return
					}
					select {
					case output <- items:
					case <-runCtx.Done():
						return
					}
				case <-runCtx.Done():
					return
				}
			}
		}()

		return inputMap[selector]
	}

	flush := func() {
		// beforeFlush()
		wg := sync.WaitGroup{}
		wg.Add(len(flushMap))
		for _, fl := range flushMap {
			go func() {
				select {
				case fl <- func() {
					wg.Done()
				}:
				case <-runCtx.Done():
				}
			}()
		}
		wg.Wait()
		// afterFlush()
	}

	go func() {
		defer cancel()
		for {
			select {
			case item, ok := <-inputCh:
				if !ok {
					flush()
					for _, inch := range inputMap {
						close(inch)
					}
					outWg.Wait()
					close(output)
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
				case <-runCtx.Done():
				}
				afterSend(item, flush)
			case done := <-flushCh:
				flush()
				done()
			case <-runCtx.Done():
				return
			}
		}
	}()

	return output
}
