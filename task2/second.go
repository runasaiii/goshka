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



/*
The commented code below shows a data race problem.
Its when multiple goroutines perform unsynchronized operations (read, write and modify)
on a shared variable, and then it can lead to lost updates and inconsistent results.
The operations on counter are not atomic and can be interrupted by other goroutines,
resulting in a final value which is less than 1000

func main() {
	var counter int
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter++
		}()
	}

	wg.Wait()

	fmt.Printf("Final counter: %d\n", counter)
}
*/
