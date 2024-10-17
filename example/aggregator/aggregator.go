package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/Skrip42/sozu"
)

func main() {
	ch := make(chan int)
	aggregateFunc := func(a, b int) int {
		if a > b {
			return a
		}
		return b
	}

	fabric := sozu.NewAggregator(aggregateFunc)

	wg := sync.WaitGroup{}

	ctx, _ := context.WithCancel(context.Background())
	resultCh, flush := fabric.Create(ctx, ch)

	wg.Add(1)
	go func() {
		// expected
		// [4]
		// [4]
		// [5]
		for items := range resultCh {
			fmt.Println(items)
		}
		wg.Done()
	}()

	data1 := []int{1, 4, 3, 2}
	data2 := []int{0, 1, 3}
	data3 := []int{0, 1, 5}

	for _, item := range data1 {
		ch <- item
	}
	flush(ctx)
	for _, item := range data2 {
		ch <- item
	}
	flush(ctx)
	for _, item := range data3 {
		ch <- item
	}
	close(ch)

	wg.Wait()

}
