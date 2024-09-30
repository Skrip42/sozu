package sozu

import (
	"context"

	"github.com/Skrip42/sozu/internal"
	"github.com/Skrip42/sozu/internal/helper"
)

type Fabric[V any] interface {
	Create(ctx context.Context, input chan V) (<-chan []V, func(context.Context))
}

type fabric[V any] struct {
	base internal.Fabric[V]
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

	return f.base.Create(
		ctx,
		input,
		flushCh,
		func(_ V, _ func()) {},
		func(_ V, _ func()) {},
		func() {},
		func() {},
		0,
	), flush
}

func New[V any](opts ...applyOptionFunc[V]) Fabric[V] {
	fb := internal.NewSimpleFabric[V]()
	for _, opt := range opts {
		fb = opt(fb)
	}
	return &fabric[V]{fb}
}
