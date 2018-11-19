package options

import "fmt"

type WorkerDatastore struct {
	Type string
	Name string
	Tags []string

	Scripts struct{
		Pre string
		Post string
	}

	Params map[string]interface{} //ds specific
}

type TestSpec struct {
	Datastore WorkerDatastore

	Test    string
	Options BenchOptions
}

type BenchOptions struct {
	PrimeRecordCount int // number of records in the datastore before the test
	RecordSize       int // size of one record
	BatchSize        int // size of the batch, only applies to batched operations
}

func (opt BenchOptions) TestDesc() string {
	return fmt.Sprintf("pre=%d-size=%d-batch=%d", opt.PrimeRecordCount, opt.RecordSize, opt.BatchSize)
}

var DefaultBenchOpts = OptionsRange2pow(
	BenchOptions{1, 25 << 10, 64},
	BenchOptions{1 << 18, 25 << 10, 64}, 8)

  // 20 / 25 / 13 is 1M / ~20G

func OptionsRange2pow(start, end BenchOptions, countPerAxis int) []BenchOptions {
	res := []BenchOptions{start}

	if start == end {
		return res
	}

	axis := make([]float64, countPerAxis)
	maxN := 1 << uint(countPerAxis)
	for i := 0; i < countPerAxis-1; i++ {
		axis[i] = float64(uint(1)<<uint(i+1)) / float64(maxN)
	}
	axis[countPerAxis-1] = 1

	if start.PrimeRecordCount != end.PrimeRecordCount {
		bRes := res[:]
		res = make([]BenchOptions, 0, countPerAxis*len(bRes))
		for _, opt := range bRes {
			for _, scale := range axis {
				opt.PrimeRecordCount = int(float64(end.PrimeRecordCount-start.PrimeRecordCount)*scale) + start.PrimeRecordCount
				res = append(res, opt)
			}
		}

	}
	if start.RecordSize != end.RecordSize {
		bRes := res[:]
		res = make([]BenchOptions, 0, countPerAxis*len(bRes))
		for _, opt := range bRes {
			for _, scale := range axis {
				opt.RecordSize = int(float64(end.RecordSize-start.RecordSize)*scale) + start.RecordSize
				res = append(res, opt)
			}
		}

	}

	if start.BatchSize != end.BatchSize {
		bRes := res[:]
		res = make([]BenchOptions, 0, countPerAxis*len(bRes))
		for _, opt := range bRes {
			for _, scale := range axis {
				opt.BatchSize = int(float64(end.BatchSize-start.BatchSize)*scale) + start.BatchSize
				res = append(res, opt)
			}
		}

	}

	return res
}
