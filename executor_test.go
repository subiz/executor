package executor

import (
	"math"
	"testing"
	"time"
)

// the executer should evenly dispatch the jobs to all goroutines.
func TestEvenly(t *testing.T) {
	maxWorkers := 5

	executor := NewExecutor(uint(maxWorkers), 10000, func(job Job) {})

	totalJob := 5000

	for i := 1; i <= totalJob; i++ {
		executor.AddJob(Job{Key: intToStr(i), Data: i})
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

	executor := NewExecutor(10, 100, func(job Job) {
		time.Sleep(100 * time.Millisecond)

		if job.Data.(string) == "2" {
			done <- true
		}
	})

	go func() {
		executor.AddJob(Job{Key: "k1", Data: "1"})
		executor.AddJob(Job{Key: "k1", Data: "2"})
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

	executor := NewExecutor(10, 100, func(job Job) {
		if job.Key == "5" {
			done = true
		} else {
			time.Sleep(1 * time.Hour)
		}
	})

	go func() {
		for i := 1; i <= 5; i++ {
			executor.AddJob(Job{Key: intToStr(i), Data: i})
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

// the executer must stop adding new job (block) if one of the queue is full.
// khi gửi (MaxJobs + 1) jobs, job thứ MaxJobs + 1 bị block
func TestBlockNewJob(t *testing.T) {
	maxJobs := 4
	executor := NewExecutor(1, uint(maxJobs), func(job Job) {
		time.Sleep(1 * time.Hour)
	})

	donec, nextc := make(chan bool, 0), make(chan bool, 0)
	go func() {
		for i := 0; i < maxJobs; i++ {
			executor.AddJob(Job{Key: "k", Data: i})
		}
		nextc <- true // this should get call

		executor.AddJob(Job{Key: "k", Data: maxJobs}) // this should block
		donec <- true
	}()

	<-nextc
	select {
	case <-donec:
		t.FailNow()
	case <-time.After(1 * time.Second):
		return
	}
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
	executor := NewExecutor(10, 10, func(job Job) {
		return
	})

	for i := 1; i <= 10; i++ {
		executor.AddJob(Job{Key: intToStr(i), Data: i})
	}

	executor.Wait()
}
