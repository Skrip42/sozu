package internal

//go:generate mockgen -typed -destination=factory_mock.go -source=factory.go -package=internal

import "context"

type Factory[V any] interface {
	Create(
		ctx context.Context,
		inputCh chan V,
		flushCh chan func(),
		beforeSend func(V, func()),
		afterSend func(V, func()),
		beforeFlush func(),
		afterFlush func(int),
		capacity int,
		cancel context.CancelFunc,
	) <-chan []V
}
