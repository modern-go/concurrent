package concurrent

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
	"runtime/debug"
)

// LogInfo logs informational message, for example
// which goroutine is still alive preventing the executor shutdown
var LogInfo = func(event string, properties ...interface{}) {
}

// LogPanic logs goroutine panic
var LogPanic = func(recovered interface{}, properties ...interface{}) interface{} {
	fmt.Println(fmt.Sprintf("paniced: %v", recovered))
	debug.PrintStack()
	return recovered
}

// StopSignal will not be recovered, will propagate to upper level goroutine
const StopSignal = "STOP!"

// UnboundedExecutor is a executor without limits on counts of alive goroutines
// it tracks the goroutine started by it, and can cancel them when shutdown
type UnboundedExecutor struct {
	ctx                   context.Context
	cancel                context.CancelFunc
	activeGoroutinesMutex *sync.Mutex
	activeGoroutines      map[string]int
}

// GlobalUnboundedExecutor has the life cycle of the program itself
// any goroutine want to be shutdown before main exit can be started from this executor
// GlobalUnboundedExecutor expects the main function to call stop
// it does not magically knows the main function exits
var GlobalUnboundedExecutor = NewUnboundedExecutor()

// NewUnboundedExecutor creates a new UnboundedExecutor,
// UnboundedExecutor can not be created by &UnboundedExecutor{}
func NewUnboundedExecutor() *UnboundedExecutor {
	ctx, cancel := context.WithCancel(context.TODO())
	return &UnboundedExecutor{
		ctx:                   ctx,
		cancel:                cancel,
		activeGoroutinesMutex: &sync.Mutex{},
		activeGoroutines:      map[string]int{},
	}
}

// Go starts a new goroutine and tracks its lifecycle.
// Panic will be recovered and logged automatically, except for StopSignal
func (executor *UnboundedExecutor) Go(handler func(ctx context.Context)) {
	_, file, line, _ := runtime.Caller(1)
	executor.activeGoroutinesMutex.Lock()
	defer executor.activeGoroutinesMutex.Unlock()
	startFrom := fmt.Sprintf("%s:%d", file, line)
	executor.activeGoroutines[startFrom] += 1
	go func() {
		defer func() {
			recovered := recover()
			if recovered != nil && recovered != StopSignal {
				LogPanic(recovered)
			}
			executor.activeGoroutinesMutex.Lock()
			defer executor.activeGoroutinesMutex.Unlock()
			executor.activeGoroutines[startFrom] -= 1
		}()
		handler(executor.ctx)
	}()
}

// Stop cancel all goroutines started by this executor without wait
func (executor *UnboundedExecutor) Stop() {
	executor.cancel()
}

// Stop cancel all goroutines started by this executor and
// wait until all goroutines exited
func (executor *UnboundedExecutor) StopAndWaitForever() {
	executor.StopAndWait(context.Background())
}

// Stop cancel all goroutines started by this executor and wait.
// Wait can be cancelled by the context passed in.
func (executor *UnboundedExecutor) StopAndWait(ctx context.Context) {
	executor.cancel()
	for {
		fiveSeconds := time.NewTimer(time.Millisecond * 100)
		select {
		case <-fiveSeconds.C:
		case <-ctx.Done():
			return
		}
		if executor.checkGoroutines() {
			return
		}
	}
}

func (executor *UnboundedExecutor) checkGoroutines() bool {
	executor.activeGoroutinesMutex.Lock()
	defer executor.activeGoroutinesMutex.Unlock()
	for startFrom, count := range executor.activeGoroutines {
		if count > 0 {
			LogInfo("event!unbounded_executor.still waiting goroutines to quit",
				"startFrom", startFrom,
				"count", count)
			return false
		}
	}
	return true
}
