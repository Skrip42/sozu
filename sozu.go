package sozu

import (
	"context"

	"github.com/Skrip42/sozu/internal"
	"github.com/Skrip42/sozu/internal/helper"
)

type Factory[V any] interface {
	Create(ctx context.Context, input chan V) (<-chan []V, func(context.Context))
}

type fabric[V any] struct {
	base internal.Factory[V]
}

func (f *fabric[V]) Create(ctx context.Context, input chan V) (<-chan []V, func(context.Context)) {
	flushCh := make(chan func())

	flush := func(ctx context.Context) {
		await, done := helper.NewSyncer(ctx)
		select {
		case flushCh <- done:
		case <-ctx.Done():
		}
		await()
	}

	innerCtx, cancel := context.WithCancel(ctx)

	return f.base.Create(
		innerCtx,
		input,
		flushCh,
		func(_ V, _ func()) {},
		func(_ V, _ func()) {},
		func() {},
		func(_ int) {},
		0,
		cancel,
	), flush
}

func NewBuffer[V any](opts ...applyOptionFunc[V]) Factory[V] {
	fb := internal.NewBufferFactory[V]()
	for _, opt := range opts {
		fb = opt(fb)
	}
	return &fabric[V]{fb}
}

func NewAggregator[V any](aggregateFunc func(V, V) V, opts ...applyOptionFunc[V]) Factory[V] {
	fb := internal.NewAggregatorFactory(aggregateFunc)
	for _, opt := range opts {
		fb = opt(fb)
	}
	return &fabric[V]{fb}
}
