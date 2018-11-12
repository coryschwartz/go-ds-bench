package worker

import (
	"github.com/ipfs/go-ds-bench/options"
	"testing"
)

func TestOptionsSimpleRange(t *testing.T) {
	start := options.BenchOptions{1, 100, 64}
	end := options.BenchOptions{1 << 10, 100, 64}

	opts := options.OptionsRange2pow(start, end, 11)
	if len(opts) != 11 {
		t.Fatalf("length is %d, should be %d", len(opts), 11)
	}

	for k, v := range opts[1:9] {
		if 1<<uint(k) != v.PrimeRecordCount {
			t.Errorf("expected PrimeRecordCount=%d, got %d @%d", 1<<uint(k), v.PrimeRecordCount, k)
		}
	}
}
func TestOptionsBoth(t *testing.T) {
	start := options.BenchOptions{1, 1, 64}
	end := options.BenchOptions{1 << 10, 1 << 10, 64}

	opts := options.OptionsRange2pow(start, end, 11)
	if len(opts) != 11*11 {
		t.Fatalf("length is %d, should be %d", len(opts), 11*11)
	}
}
