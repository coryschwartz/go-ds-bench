package worker

import (
	"github.com/ipfs/go-ds-bench/options"
	"testing"

	ds "github.com/ipfs/go-datastore"
)

func BenchGetAt(b *testing.B, store ds.Batching, opt options.BenchOptions) {
	PrimeDS(b, store, opt.PrimeRecordCount, opt.RecordSize)
	buf := make([]byte, opt.RecordSize)
	keys := make([]ds.Key, b.N)
	for i := 0; i < b.N; i++ {
		buf = RandomBuf(opt.RecordSize)
		keys[i] = ds.RandomKey()

		store.Put(keys[i], buf)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := store.Get(keys[i])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchGetSeries(b *testing.B, newStore CandidateDatastore, opts []options.BenchOptions) {
	for _, opt := range opts {
		b.Run(opt.TestDesc(), func(b *testing.B) {
			store, err := newStore.Create()
			if err != nil {
				b.Fatal(err)
			}
			BenchGetAt(b, store, opt)

			b.StopTimer()
			newStore.Destroy(store)
		})
	}
}
