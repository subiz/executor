package executor

import (
	"sync"
	"testing"
	"time"
)

func TestAsync(t *testing.T) {
	arr := []int{1, 2, 4, 4, 5}
	result := make([]int, len(arr), len(arr))
	Async(len(arr), func(i int, lock *sync.Mutex) {
		result[i] = arr[i] * 2
		time.Sleep(time.Duration((i%3)+1) * time.Second)
	}, 3)

	Async(len(arr), func(i int, lock *sync.Mutex) {
		result[i] = arr[i] * 2
	}, 10)

	Async(0, func(i int, lock *sync.Mutex) {
		result[i] = arr[i] * 2
	}, 10)

	Async(len(arr), func(i int, lock *sync.Mutex) {
		result[i] = arr[i] * 2
	}, 0)

	Async(-1, func(i int, lock *sync.Mutex) {
		result[i] = arr[i] * 2
	}, 0)

	Async(0, func(i int, lock *sync.Mutex) {
		result[i] = arr[i] * 2
	}, -1)

	Async(-1, func(i int, lock *sync.Mutex) {
		result[i] = arr[i] * 2
	}, -1)
}
