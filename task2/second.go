package main
import (
	"fmt"
	"sync"
	"sync/atomic"
)


func withMutex() {
	var counter int
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	fmt.Println("Counter with mutex:", counter)
}

func withAtomic() {
	var counter atomic.Int64 
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Add(1)
		}()
	}
	wg.Wait()
	fmt.Println("Counter with atomic:", counter.Load())
}

func main() {
	withMutex()
	withAtomic()
}