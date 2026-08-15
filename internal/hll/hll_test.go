package hll

import "testing"

func TestSmallCardinality(t *testing.T) {
	h := New()
	for i := 0; i < 100; i++ {
		h.Add(string(rune('a'+i%26)) + string(rune(i)))
	}
	est := h.Estimate()
	if est < 90 || est > 110 {
		t.Fatalf("estimate=%v", est)
	}
}
