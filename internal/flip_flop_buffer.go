package internal

import "context"

type flipFlopFabric[V any, C comparable] struct {
	base     Fabric[V]
	criteria func(V) C
}

func NewFlipFlopFabric[V any, C comparable](base Fabric[V], criteria func(V) C) Fabric[V] {
	return &flipFlopFabric[V, C]{base: base, criteria: criteria}
}

func (f *flipFlopFabric[V, C]) Create(
	ctx context.Context,
	inputCh chan V,
	flushCh chan func(),
	beforeSend func(V, func()),
	afterSend func(V, func()),
	beforeFlush func(),
	afterFlush func(int),
	capacity int,
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
	)
}
