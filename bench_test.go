package main

import "testing"

func BenchmarkAddBatchSeriesDefault(b *testing.B) {
	for _, c := range AllCandidates {
		BenchAddBatchSeries(b, c, DefaultBenchOpts)
	}
}

func BenchmarkAddSeriesDefault(b *testing.B) {
	for _, c := range AllCandidates {
		BenchAddSeries(b, c, DefaultBenchOpts)
	}
}

func BenchmarkAddBatchAll(b *testing.B) {
	for _, c := range AllCandidates {
		BenchAddBatchSeries(b, c, []BenchOptions{{0, 256 << 10, 128}})
	}
}

func BenchmarkGetSeriesDefault(b *testing.B) {
	BenchGetSeries(b, CandidateMemoryMap, DefaultBenchOpts)
}

func BenchmarkHasSeriesDefault(b *testing.B) {
	BenchHasSeries(b, CandidateMemoryMap, DefaultBenchOpts)
}