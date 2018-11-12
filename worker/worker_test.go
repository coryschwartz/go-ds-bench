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
		BenchGetSeries(b, CandidateDs(spec.Datastore), []options.BenchOptions{spec.Options})
	case "has":
		BenchHasSeries(b, CandidateDs(spec.Datastore), []options.BenchOptions{spec.Options})
	case "add":
		BenchAddSeries(b, CandidateDs(spec.Datastore), []options.BenchOptions{spec.Options})
	case "add-batch":
		BenchAddBatchSeries(b, CandidateDs(spec.Datastore), []options.BenchOptions{spec.Options})
	default:
		b.Fatalf("unknown test '%s'", spec.Test)
	}
}
