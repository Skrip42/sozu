package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/Skrip42/sozu"
)

func main() {
	ch := make(chan int)

	fabric := sozu.NewBuffer[int](sozu.WithLimit[int](6))

	wg := sync.WaitGroup{}

	ctx, _ := context.WithCancel(context.Background())
	resultCh, flush := fabric.Create(ctx, ch)

	wg.Add(1)
	go func() {
		for items := range resultCh {
			// expected
			// [0 1 2 3 4 5]
			// [6 7]
			// [8 9 10 11]
			fmt.Println(items)
		}
		wg.Done()
	}()

	for i := 0; i < 8; i++ {
		ch <- i
	}
	flush(ctx)
	for i := 8; i < 12; i++ {
		ch <- i
	}
	close(ch)

	wg.Wait()
}
