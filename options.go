package sozu

import (
	"time"

	"github.com/Skrip42/sozu/internal"
)

type applyOptionFunc[V any] func(internal.Factory[V]) internal.Factory[V]

func WithLimit[V any](limit int) applyOptionFunc[V] {
	return func(base internal.Factory[V]) internal.Factory[V] {
		return internal.NewLimitedFactory(base, limit)
	}
}

func WithTimer[V any](duratiorn time.Duration) applyOptionFunc[V] {
	return func(base internal.Factory[V]) internal.Factory[V] {
		return internal.NewTimerFactory(base, duratiorn)
	}
}

func WithFlipFlop[V any, C comparable](criteria func(V) C) applyOptionFunc[V] {
	return func(base internal.Factory[V]) internal.Factory[V] {
		return internal.NewFlipFlopFactory(base, criteria)
	}
}

func WithMultiplexor[V any, C comparable](
	separator func(V) C,
	capacity int,
) applyOptionFunc[V] {
	return func(base internal.Factory[V]) internal.Factory[V] {
		return internal.NewMultiplexorFactory(base, separator, capacity)
	}
}
