package sozu

import (
	"time"

	"github.com/Skrip42/sozu/internal"
)

type applyOptionFunc[V any] func(internal.Fabric[V]) internal.Fabric[V]

func WithLimit[V any](limit int) applyOptionFunc[V] {
	return func(base internal.Fabric[V]) internal.Fabric[V] {
		return internal.NewLimitedFabric(base, limit)
	}
}

func WithTimer[V any](duratiorn time.Duration) applyOptionFunc[V] {
	return func(base internal.Fabric[V]) internal.Fabric[V] {
		return internal.NewTimerFabric(base, duratiorn)
	}
}

func WithFlipFlop[V any, C comparable](criteria func(V) C) applyOptionFunc[V] {
	return func(base internal.Fabric[V]) internal.Fabric[V] {
		return internal.NewFlipFlopFabric(base, criteria)
	}
}

func WithMultiplexor[V any, C comparable](
	separator func(V) C,
	capacity int,
) applyOptionFunc[V] {
	return func(base internal.Fabric[V]) internal.Fabric[V] {
		return internal.NewMultiplexorFabric(base, separator, capacity)
	}
}
