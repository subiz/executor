package executor

import (
	"sync"
	"testing"
	"time"
)

// the executer must execute 2 jobs with the same key sequencely.
func TestSequencely(t *testing.T) {
	startTime := time.Now()
	done := make(chan bool)

	exe := New(func(key string, data interface{}) {
		time.Sleep(100 * time.Millisecond)

		if data.(string) == "2" {
			done <- true
		}
	})

	go func() {
		exe.Add("k1", "1")
		exe.Add("k1", "2")
	}()

	<-done

	elapsed := time.Now().Sub(startTime)

	if elapsed < 200*time.Millisecond {
		t.Fatalf("processed time less than 200 miliseconds, got %#v", elapsed/time.Millisecond)
	}
}

// the executer should execute 2 jobs with different keys concurrently.
func TestConcurrently(t *testing.T) {
	done := false

	exe := New(func(key string, data interface{}) {
		if key == "5" {
			done = true
		} else {
			time.Sleep(1 * time.Hour)
		}
	})

	go func() {
		for i := 1; i <= 5; i++ {
			exe.Add(intToStr(i), i)
		}
	}()

	go func() {
		time.Sleep(50 * time.Millisecond)
		if !done {
			t.FailNow()
		}
	}()

	time.Sleep(100 * time.Millisecond) // wait checker
}

func TestWait(t *testing.T) {
	mu := &sync.Mutex{}
	i := 0
	exe := New(func(_ string, _ interface{}) {
		mu.Lock()
		i++
		mu.Unlock()
	})
	exe.Add(intToStr(i), i)

	exe.Wait()
	mu.Lock()

	if i != 1 {
		t.Fatal(i)
	}
	mu.Unlock()
}

func TestTearDown(t *testing.T) {
	exe := New(func(_ string, _ interface{}) {})

	go func() {
		for i := 1; i <= 100; i++ {
			exe.Add(intToStr(i), i)
		}
	}()

	exe.Wait()
	time.Sleep(TIMEOUT + 1*time.Second)

	info := exe.Info()
	if len(info) != 0 {
		t.FailNow()
	}
}
