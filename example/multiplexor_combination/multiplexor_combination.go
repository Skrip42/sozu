package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/Skrip42/sozu"
)

func main() {
	ch := make(chan int)
	input := []int{1, 2, 3, 3, 3, 3, 3, 1, 2, 3, 1, 2, 3, 1, 2}
	separator := func(i int) int {
		return i % 3
	}

	fabric := sozu.NewBuffer[int](
		sozu.WithLimit[int](5), // limit = 5 for each
		sozu.WithMultiplexor(separator, 3),
		sozu.WithLimit[int](10), // limit = 10 for all
	)

	wg := sync.WaitGroup{}

	ctx, _ := context.WithCancel(context.Background())
	resultCh, flush := fabric.Create(ctx, ch)

	wg.Add(1)
	go func() {
		for items := range resultCh {
			// expected
			// [ 3, 3, 3, 3, 3] // trigger some buffer
			// [1, 1, 1, 1] // trigger all buffer
			// [ 2, 2, 2, 2]
			// [ 3, 3]
			fmt.Println(items)
		}
		wg.Done()
	}()

	for _, i := range input {
		ch <- i
	}
	flush(ctx)
	close(ch)

	wg.Wait()
}
