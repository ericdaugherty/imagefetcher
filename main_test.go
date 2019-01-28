package main

import (
	"testing"
	"time"
)

func TestIsAwake(t *testing.T) {

	n := time.Now()
	t1 := time.Date(n.Year(), n.Month(), n.Day(), 23, 1, 0, 0, n.Location())

	if isAwake(t1, 22, 7) {
		t.Error("For hour 23, should be asleep but is awake.")
	}

	t2 := time.Date(n.Year(), n.Month(), n.Day(), 21, 59, 0, 0, n.Location())

	if !isAwake(t2, 22, 7) {
		t.Error("For hour 21, should be awake but is asleep.")
	}

	t3 := time.Date(n.Year(), n.Month(), n.Day(), 7, 1, 0, 0, n.Location())

	if !isAwake(t3, 22, 7) {
		t.Error("For hour 7, should be awake but is asleep.")
	}

	t10 := time.Date(n.Year(), n.Month(), n.Day(), 7, 1, 0, 0, n.Location())

	if !isAwake(t10, 1, 7) {
		t.Error("For hour 7, should be awake but is asleep.")
	}

	t11 := time.Date(n.Year(), n.Month(), n.Day(), 0, 1, 0, 0, n.Location())

	if !isAwake(t11, 1, 7) {
		t.Error("For hour 0, should be awake but is asleep.")
	}

	t12 := time.Date(n.Year(), n.Month(), n.Day(), 4, 1, 0, 0, n.Location())

	if isAwake(t12, 1, 7) {
		t.Error("For hour 4, should be asleep but is awake.")
	}
}
