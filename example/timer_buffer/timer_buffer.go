package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Skrip42/sozu"
)

func main() {
	ch := make(chan int)

	fabric := sozu.NewBuffer[int](sozu.WithTimer[int](time.Millisecond * 100))

	wg := sync.WaitGroup{}

	ctx, _ := context.WithCancel(context.Background())
	resultCh, flush := fabric.Create(ctx, ch)

	wg.Add(1)
	go func() {
		for items := range resultCh {
			// expected
			// [0 1 2 3 4 5 6 7 8 9]
			// [10 11 12 13 14 15]
			// [16 17]
			fmt.Println(items)
		}
		wg.Done()
	}()

	for i := 0; i < 16; i++ {
		ch <- i
		time.Sleep(time.Millisecond * 10)
	}

	flush(ctx)
	ch <- 16
	ch <- 17
	close(ch)

	wg.Wait()
}
