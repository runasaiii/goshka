package main
import (
	"fmt"
	"sync"
)

type SafeMap struct {
	mu sync.RWMutex
	m  map[string]int
}

func (s *SafeMap) Set(key string, val int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = val
}

func (s *SafeMap) Get(key string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.m[key]
}

func main() {
	var wg sync.WaitGroup
	
	sm := &SafeMap{m: make(map[string]int)}
	
	var sMap sync.Map

	iterations := 100
	wg.Add(iterations * 2)

	for i := 0; i < iterations; i++ {
		go func(v int) {
			defer wg.Done()
			sm.Set("key", v)
		}(i)

		go func(v int) {
			defer wg.Done()
			sMap.Store("key", v)
		}(i)
	}

	wg.Wait()

	fmt.Printf("RWMutex map value is: %d\n", sm.Get("key"))
	
	val, _ := sMap.Load("key")
	fmt.Printf("sync.Map value is: %v\n", val)
}