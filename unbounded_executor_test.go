package concurrent_test

import (
	"context"
	"fmt"
	"time"
	"github.com/modern-go/concurrent"
)

func ExampleUnboundedExecutor_Go() {
	executor := concurrent.NewUnboundedExecutor()
	executor.Go(func(ctx context.Context) {
		fmt.Println("abc")
	})
	time.Sleep(time.Second)
	// output: abc
}

func ExampleUnboundedExecutor_StopAndWaitForever() {
	executor := concurrent.NewUnboundedExecutor()
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

func ExampleUnboundedExecutor_Go_panic() {
	concurrent.HandlePanic = func(recovered interface{}, file string, line int) {
		fmt.Println("panic logged")
	}
	executor := concurrent.NewUnboundedExecutor()
	executor.Go(func(ctx context.Context) {
		panic("!!!")
	})
	time.Sleep(time.Second)
	// output: panic logged
}
