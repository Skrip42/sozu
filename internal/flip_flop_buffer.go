package internal

import "context"

type flipFlopFactory[V any, C comparable] struct {
	base     Factory[V]
	criteria func(V) C
}

func NewFlipFlopFactory[V any, C comparable](base Factory[V], criteria func(V) C) Factory[V] {
	return &flipFlopFactory[V, C]{base: base, criteria: criteria}
}

func (f *flipFlopFactory[V, C]) Create(
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
	var state C
	isStateInitialize := false

	send := func(item V, flush func()) {
		beforeSend(item, flush)
		itemState := f.criteria(item)
		if !isStateInitialize {
			state = itemState
			isStateInitialize = true
		}
		if state != itemState {
			flush()
			state = itemState
		}
	}

	return f.base.Create(
		ctx,
		inputCh,
		flushCh,
		send,
		afterSend,
		beforeFlush,
		afterFlush,
		capacity,
		cancel,
	)
}
