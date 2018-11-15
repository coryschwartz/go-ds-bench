package worker

import (
	"syscall"
	"testing"

	"github.com/ipfs/go-ds-bench/options"

	ds "github.com/ipfs/go-datastore"
)

type BenchFunc func(b *testing.B, store ds.Batching, opt options.BenchOptions)

func RunBench(b *testing.B, bf BenchFunc, store CandidateDatastore, opt options.BenchOptions) {
	b.Run(opt.TestDesc(), func(b *testing.B) {
		newStore, err := store.Create()
		if err != nil {
			b.Fatal(err)
		}

		s, closer, err := newStore(true)
		if err != nil {
			b.Fatal(err)
		}
		PrimeDS(b, s, opt.PrimeRecordCount, opt.RecordSize)
		closer.Close()
		syscall.Sync()

		s, closer, err = newStore(false)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		bf(b, s, opt)
		b.StopTimer()

		closer.Close()
		store.Destroy()
	})
}
