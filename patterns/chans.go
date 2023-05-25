package patterns

import "sync"

func Merge[T any](cs ...<-chan T) <-chan T {
	var wg sync.WaitGroup

	out := make(chan T)

	send := func(c <-chan T) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go send(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func Split[T any](inputCh <-chan T, n int) []chan T {
	outputChannels := make([]chan T, 0)
	for i := 0; i < n; i++ {
		outputChannels = append(outputChannels, make(chan T))
	}

	go func(inputCh <-chan T, outputChs []chan T) {
		defer func(cs []chan T) {
			for _, c := range cs {
				close(c)
			}
		}(outputChs)

		for {
			for _, c := range outputChs {
				select {
				case val, ok := <-inputCh:
					if !ok {
						return
					}

					c <- val
				}
			}
		}
	}(inputCh, outputChannels)

	return outputChannels
}
