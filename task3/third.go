package main
import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)


func startServer(ctx context.Context, name string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(rand.Intn(500)) * time.Millisecond):
				out <- fmt.Sprintf("[%s] metric: %d", name, rand.Intn(100))
			}
		}
	}()
	return out
}

func FanIn(ctx context.Context, channels ...<-chan string) <-chan string {
    finalOut := make(chan string)
    var wg sync.WaitGroup

    multiplex := func(ch <-chan string) {
        defer wg.Done()
        for {
            select {
            case <-ctx.Done():
                return
            case val, ok := <-ch:
                if !ok {
                    return
                }
                select {
                case finalOut <- val:
                case <-ctx.Done():
                    return
                }
            }
        }
    }

    wg.Add(len(channels))
    for _, c := range channels {
        go multiplex(c)
    }

    go func() {
        wg.Wait()
        close(finalOut)
    }()

    return finalOut
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch1 := startServer(ctx, "Alpha")
	ch2 := startServer(ctx, "Beta")
	ch3 := startServer(ctx, "Gamma")

	mergedCh := FanIn(ctx, ch1, ch2, ch3)

	for val := range mergedCh {
		fmt.Println(val)
	}
}