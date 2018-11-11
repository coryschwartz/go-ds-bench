package main

import "testing"

func BenchmarkAddSeriesDefault(b *testing.B) {
	for _, c := range AllCandidates {
		BenchAddSeriesDefault(b, c)
	}
}
