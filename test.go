package main

import (
	"fmt"
	_ "time"
	"sync"
	"context"
)

var wg sync.WaitGroup

func generator(ctx context.Context, num int) <-chan int {
	wg.Add(1)
	out := make(chan int)
	go func() {
		defer wg.Done()
	Loop:
		for {
			select {
			case <-ctx.Done():
				break Loop
			case out <- num:
			}
		}
		close(out)
		userID, authToken, tracID := ctx.Value("userID").(int), ctx.Value("authToken").(string), ctx.Value("tracID").(string)
		fmt.Println("log: ", userID, authToken, tracID)
		fmt.Println("generator done")
	}()
	return out
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	ctx = context.WithValue(ctx, "userID", 123)
	ctx = context.WithValue(ctx, "authToken", "token123")
	ctx = context.WithValue(ctx, "tracID", "trace123")
	gen := generator(ctx, 1)

	for i := 0; i < 5; i++ {
		fmt.Println(<-gen)
	}
	cancel()
	wg.Wait()


}