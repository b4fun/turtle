package turtle

import (
	"context"
	"sync"
)

func runWithWorker(ctx context.Context, n int, f func(ctx context.Context, workerId int)) {
	var wg sync.WaitGroup
	work := make(chan int, n)

	for i := 0; i < n; i++ {
		work <- i
	}

	spawn := func(workerId int) {
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				work <- workerId
			}()

			f(ctx, workerId)
		}()
	}

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case workerId := <-work:
			spawn(workerId)
		}
	}

	wg.Wait()
}
