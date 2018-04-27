package executor

import (
	"testing"
)

func TestGetWorkerID(t *testing.T) {
	var maxWorkers uint = 10
	key := "k1"
	expected := getWorkerID(key, maxWorkers)
	got := getWorkerID(key, maxWorkers)

	if expected != got {
		t.Errorf("expected %d, got: %d", expected, got)
	}
}
