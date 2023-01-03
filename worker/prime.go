package worker

import (
	"context"
	"math/rand"
	"sync"
	"testing"

	ds "github.com/ipfs/go-datastore"
)

const primeMaxBatchSize = 1 << 30 // 1 GiB

func PrimeDS(tb testing.TB, store ds.Batching, count, blockSize int) {
	maxBatchCount := primeMaxBatchSize / blockSize
	if maxBatchCount > 2048 {
		maxBatchCount = 2048
	}
	// parallelism := 256
	// if _, ok := store.(ds.ThreadSafeDatastore); !ok {
	// 	parallelism = 1
	// }
	parallelism := 1
	maxBatchCount = maxBatchCount / parallelism

	var wg sync.WaitGroup
	wg.Add(parallelism)
	for i := 0; i < parallelism; i++ {
		go func() {
			defer wg.Done()
			ctx := context.Background()
			buf := make([]byte, blockSize)
			b, err := store.Batch(ctx)
			if err != nil {
				tb.Fatal(err)
			}

			for i := 0; i < count/parallelism; i++ {
				_, err := rand.Read(buf)
				if err != nil {
					tb.Fatal(err)
				}
				err = b.Put(ctx, ds.RandomKey(), buf)
				if err != nil {
					tb.Fatal(err)
				}

				if i%maxBatchCount == maxBatchCount-1 {
					err = b.Commit(ctx)
					if err != nil {
						tb.Fatal(err)
					}
					b, err = store.Batch(ctx)
					if err != nil {
						tb.Fatal(err)
					}
				}
			}

			err = b.Commit(ctx)
			if err != nil {
				tb.Fatal(err)
			}
		}()
	}
	wg.Wait()
}
