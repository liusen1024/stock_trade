package mw

import (
	"fmt"
	"testing"
	"time"
)

func TestLFSR(t *testing.T) {
	f := &FibonacciLFSR{
		state: 44257,
	}
	if f.Next() != 22128 {
		t.Fatal()
	}
}

func TestRandomSecret(t *testing.T) {
	tm := time.Date(1921, 7, 25, 0, 1, 4, 0, time.Local)
	r := RandomSecret(tm)
	if r != 34391 {
		t.Fatal()
	}
	fmt.Printf("secret = %d\n", r)
}
