package internal

import "context"

type limitedFactory[V any] struct {
	base  Factory[V]
	limit int
}

func NewLimitedFactory[V any](base Factory[V], limit int) Factory[V] {
	return &limitedFactory[V]{base: base, limit: limit}
}

func (f *limitedFactory[V]) Create(
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

	send := func(item V, flush func()) {
		counter++
		if counter >= f.limit {
			flush()
		}
		afterSend(item, flush)
	}

	flush := func(flushCount int) {
		counter -= flushCount
		afterFlush(flushCount)
	}

	return f.base.Create(
		ctx,
		inputCh,
		flushCh,
		beforeSend,
		send,
		beforeFlush,
		flush,
		f.limit,
		cancel,
	)
}
