package worker

import (
	"github.com/ipfs/go-ds-bench/options"
	"testing"

	ds "github.com/ipfs/go-datastore"
)

func BenchGet(b *testing.B, store ds.Batching, opt options.BenchOptions) {
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