package worker

import (
	"github.com/ipfs/go-ds-bench/options"
	"github.com/remeh/sizedwaitgroup"
	"testing"

	ds "github.com/ipfs/go-datastore"
)

func BenchHas(b *testing.B, store ds.Batching, opt options.BenchOptions) {
	buf := make([]byte, opt.RecordSize)
	keys := make([]ds.Key, b.N)
	swg := sizedwaitgroup.New(256)

	for i := 0; i < b.N; i++ {
		buf = RandomBuf(opt.RecordSize)
		keys[i] = ds.RandomKey()

		if i%2 == 0 {
			swg.Add()
			go func(i int) {
				swg.Done()
				store.Put(keys[i], buf)
			}(i)
		}
	}
	swg.Wait()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := store.Has(keys[i])
		if err != nil {
			b.Fatal(err)
		}
	}
}
