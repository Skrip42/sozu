package internal

import (
	"context"
)

type simpleFabric[V any] struct {
}

func NewSimpleFabric[V any]() Fabric[V] {
	return &simpleFabric[V]{}
}

func (f *simpleFabric[V]) Create(
	ctx context.Context,
	inputCh chan V,
	flushCh chan func(),
	beforeSend func(V, func()),
	afterSend func(V, func()),
	beforeFlush func(),
	afterFlush func(),
	capacity int,
) <-chan []V {
	output := make(chan []V)
	buffer := make([]V, 0, capacity)

	flush := func() {
		if len(buffer) == 0 {
			return
		}
		beforeFlush()
		tmp := make([]V, len(buffer))
		copy(tmp, buffer)
		buffer = buffer[:0]

		select {
		case output <- tmp:
		case <-ctx.Done():
		}

		afterFlush()
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
			case done := <-flushCh:
				flush()
				done()
			case <-ctx.Done():
			}
		}
	}()

	return output
}
