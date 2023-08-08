package turtle

import (
	"context"
	"sync"
)

func runWithWorker(ctx context.Context, n int, f func(ctx context.Context)) {
	var wg sync.WaitGroup
	work := make(chan struct{}, n)

	for i := 0; i < n; i++ {
		work <- struct{}{}
	}

	spawn := func() {
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				work <- struct{}{}

				if r := recover(); r != nil {
					// TODO: log error
				}
			}()

			f(ctx)
		}()
	}

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-work:
			spawn()
		}
	}

	wg.Wait()
}
