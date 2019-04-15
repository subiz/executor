package executor

import (
	"math"
	"sync"
	"testing"
	"time"
)

// the executer should evenly dispatch the jobs to all goroutines.
func TestEvenly(t *testing.T) {
	maxWorkers := 5

	executor := New(uint(maxWorkers), func(_ string, _ interface{}) {})

	totalJob := 50000

	for i := 1; i <= totalJob; i++ {
		executor.Add(intToStr(i), i)
	}

	mean := totalJob / maxWorkers
	samples := []int{}

	info := executor.Info()
	for _, counter := range info {
		samples = append(samples, int(counter))
	}

	stdDeviation := getStdDeviation(samples, mean)

	expected := 0.05
	got := stdDeviation / float64(mean)
	if got >= expected {
		t.Errorf("expected < %#v, got: %#v", expected, got)
	}
}

// the executer must execute 2 jobs with the same key sequencely.
func TestSequencely(t *testing.T) {
	startTime := time.Now()
	done := make(chan bool)

	executor := New(10, func(key string, data interface{}) {
		time.Sleep(100 * time.Millisecond)

		if data.(string) == "2" {
			done <- true
		}
	})

	go func() {
		executor.Add("k1", "1")
		executor.Add("k1", "2")
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

	executor := New(10, func(key string, data interface{}) {
		if key == "5" {
			done = true
		} else {
			time.Sleep(1 * time.Hour)
		}
	})

	go func() {
		for i := 1; i <= 5; i++ {
			executor.Add(intToStr(i), i)
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

func getStdDeviation(samples []int, mean int) float64 {
	sum := 0

	for _, v := range samples {
		deviation := v - mean
		sum += deviation * deviation
	}

	variance := sum / (len(samples) - 1)
	stdDeviation := math.Sqrt(float64(variance))

	return stdDeviation
}

func TestWait(t *testing.T) {
	mu := &sync.Mutex{}
	i := 0
	executor := New(10, func(_ string, _ interface{}) {
		mu.Lock()
		i++
		mu.Unlock()
	})
	executor.Add(intToStr(i), i)

	executor.Wait()
	mu.Lock()

	if i != 1 {
		t.Fatal(i)
	}
	mu.Unlock()
}
