package executor

import (
	"sync"
)

func Async(N int, f func(i int, lock *sync.Mutex), limit int) {
	lock := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	if N <= 0 {
		return
	}
	if limit < 1 {
		limit = 1
	}
	if limit > N {
		limit = N
	}
	wg.Add(limit)
	for n := 0; n < limit; n++ {
		go func(n int) {
			for i := n; i < N; i += limit {
				f(i, lock)
			}
			wg.Done()
		}(n)
	}
	wg.Wait()
}
