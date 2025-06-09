package internal

import (
	"context"
)

type aggregatorFactory[V any] struct {
	aggregateFunc func(a, b V) V
}

func NewAggregatorFactory[V any](
	aggregateFunc func(a, b V) V,
) Factory[V] {
	return &aggregatorFactory[V]{
		aggregateFunc: aggregateFunc,
	}
}

func (f *aggregatorFactory[V]) Create(
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
	counter := 0
	var aggregateValue V

	flush := func() {
		if counter == 0 {
			return
		}
		beforeFlush()
		flushCount := counter
		counter = 0

		output <- []V{aggregateValue}

		afterFlush(flushCount)
	}

	go func() {
		defer cancel()
		defer close(output)
		for {
			select {
			case item, ok := <-inputCh:
				if !ok {
					flush()
					return
				}
				beforeSend(item, flush)
				aggregateValue = f.aggregateFunc(aggregateValue, item)
				counter++
				afterSend(item, flush)
			case done, ok := <-flushCh:
				if !ok {
					flush()
					return
				}
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
