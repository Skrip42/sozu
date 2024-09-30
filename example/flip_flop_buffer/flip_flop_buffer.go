package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/Skrip42/sozu"
)

func main() {
	ch := make(chan int)

	input := []int{1, 3, 5, 2, 4, 6, 8, 10, 7, 9}
	criteria := func(i int) bool {
		return i%2 == 0
	}

	fabric := sozu.New[int](sozu.WithFlipFlop(criteria))

	wg := sync.WaitGroup{}

	ctx, _ := context.WithCancel(context.Background())
	resultCh, flush := fabric.Create(ctx, ch)

	wg.Add(1)
	go func() {
		for items := range resultCh {
			// expected
			// [1 2 3]
			// [2 4 5 8 10]
			// [7 9]
			// [11, 13]
			fmt.Println(items)
		}
		wg.Done()
	}()

	for _, i := range input {
		ch <- i
	}
	flush(ctx)
	ch <- 11
	ch <- 13
	close(ch)

	wg.Wait()
}
