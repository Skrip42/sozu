package internal

import (
	"context"
)

type bufferFabric[V any] struct {
}

func NewBufferFabric[V any]() Fabric[V] {
	return &bufferFabric[V]{}
}

func (f *bufferFabric[V]) Create(
	ctx context.Context,
	inputCh chan V,
	flushCh chan func(),
	beforeSend func(V, func()),
	afterSend func(V, func()),
	beforeFlush func(),
	afterFlush func(int),
	capacity int,
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

		select {
		case output <- tmp:
		case <-ctx.Done():
		}

		afterFlush(flushCount)
	}

	go func() {
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
				return
			}
		}
	}()

	return output
}
