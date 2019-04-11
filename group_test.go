package executor

import (
	"testing"
	"time"
)

func TestGroup(t *testing.T) {
	eg := NewExecutorGroup(10)

	total := 0
	group := eg.NewGroup(func(key string, value interface{}) {
		total++
		time.Sleep(1 * time.Second)
	})

	for i := 0; i < 10; i++ {
		group.Add("test", i)
	}

	group.Wait()

	if total != 10 {
		t.FailNow()
	}
}
