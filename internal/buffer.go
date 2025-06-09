package internal

import (
	"context"
)

type bufferFactory[V any] struct {
}

func NewBufferFactory[V any]() Factory[V] {
	return &bufferFactory[V]{}
}

func (f *bufferFactory[V]) Create(
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
	buffer := make([]V, 0, capacity)

	flush := func() {
		if len(buffer) == 0 {
			return
		}
		beforeFlush()
		flushCount := len(buffer)
		tmp := make([]V, flushCount)
		copy(tmp, buffer)
		buffer = buffer[:0]

		output <- tmp
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
				buffer = append(buffer, item)
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
