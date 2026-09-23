package main

import "testing"

func TestSmoke(t *testing.T) {
	if got := 1 + 1; got != 2 {
		t.Fatalf("1+1 = %d, 想要 2", got)
	}
}
