package concurrent

import (
	"context"
	"fmt"
	"time"
)

func ExampleUnboundedExecutor_Go() {
	executor := NewUnboundedExecutor()
	executor.Go(func(ctx context.Context) {
		fmt.Println("abc")
	})
	time.Sleep(time.Second)
	// output: abc
}

func ExampleUnboundedExecutor_StopAndWaitForever() {
	executor := NewUnboundedExecutor()
	executor.Go(func(ctx context.Context) {
		everyMillisecond := time.NewTicker(time.Millisecond)
		for {
			select {
			case <-ctx.Done():
				fmt.Println("goroutine exited")
				return
			case <-everyMillisecond.C:
				// do something
			}
		}
	})
	time.Sleep(time.Second)
	executor.StopAndWaitForever()
	fmt.Println("exectuor stopped")
	// output:
	// goroutine exited
	// exectuor stopped
}
