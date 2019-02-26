package worker

import (
	"encoding/json"
	"github.com/ipfs/go-ds-bench/options"
	"github.com/ipfs/go-ds-bench/worker/benches/basic"
	"io/ioutil"
	"testing"
)

func BenchmarkSpec(b *testing.B) {
	j, err := ioutil.ReadFile("spec.json")
	if err != nil {
		b.Fatal(err)
	}

	var spec options.TestSpec
	if err := json.Unmarshal(j, &spec); err != nil {
		b.Fatal(err)
	}

	switch spec.Test {
	case "get":
		RunBench(b, basic.BenchGet, CandidateDs(spec.Datastore), spec.Options)
	case "has":
		RunBench(b, basic.BenchHas, CandidateDs(spec.Datastore), spec.Options)
	case "add":
		RunBench(b, basic.BenchAdd, CandidateDs(spec.Datastore), spec.Options)
	case "add-batch":
		RunBench(b, basic.BenchAddBatch, CandidateDs(spec.Datastore), spec.Options)
	default:
		b.Fatalf("unknown test '%s'", spec.Test)
	}
}
