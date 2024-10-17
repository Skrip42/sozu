package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/Skrip42/sozu"
)

func main() {
	ch := make(chan int)
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	separator := func(i int) int {
		return i % 3
	}

	fabric := sozu.NewBuffer[int](sozu.WithMultiplexor(separator, 3))

	wg := sync.WaitGroup{}

	ctx, _ := context.WithCancel(context.Background())
	resultCh, flush := fabric.Create(ctx, ch)

	wg.Add(1)
	go func() {
		for items := range resultCh {
			// expected
			// [ 1, 4, 7, 10]
			// [ 2, 5, 8, 11]
			// [ 3, 6, 9, 12]
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
