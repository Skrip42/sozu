package internal

import "context"

type Fabric[V any] interface {
	Create(
		ctx context.Context,
		inputCh chan V,
		flushCh chan func(),
		beforeSend func(V, func()),
		afterSend func(V, func()),
		beforeFlush func(),
		afterFlush func(),
		capacity int,
	) <-chan []V
}
