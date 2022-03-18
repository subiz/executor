package executor

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestAsync(t *testing.T) {
	arr := []int{1, 2, 4, 4, 5}
	result := make([]int, len(arr), len(arr))
	Async(10, func(i int, lock *sync.Mutex) {
		result[i] = arr[i] * 2
		time.Sleep(time.Duration((i%3)+1) * time.Second)
	}, 3)
}
