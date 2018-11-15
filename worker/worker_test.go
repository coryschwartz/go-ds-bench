package worker

import (
	"encoding/json"
	"github.com/ipfs/go-ds-bench/options"
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
		RunBench(b, BenchGet, CandidateDs(spec.Datastore), spec.Options)
	case "has":
		RunBench(b, BenchHas, CandidateDs(spec.Datastore), spec.Options)
	case "add":
		RunBench(b, BenchAdd, CandidateDs(spec.Datastore), spec.Options)
	case "add-batch":
		RunBench(b, BenchAddBatch, CandidateDs(spec.Datastore), spec.Options)
	default:
		b.Fatalf("unknown test '%s'", spec.Test)
	}
}
