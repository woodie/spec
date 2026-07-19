package spec_test

import (
	"testing"

	"github.com/sclevine/spec"
)

func TestAliases(t *testing.T) {
	spec.Run(t, "Aliases", func(t *testing.T, describe spec.Describe, it spec.S) {
		before, after, context := spec.Aliases(describe, it)

		var log []string
		before(func() { log = append(log, "before") })
		after(func() { log = append(log, "after") })

		context("a nested group", func() {
			it("runs before, the spec, then after", func() {
				log = append(log, "spec")
				t.Cleanup(func() {
					want := []string{"before", "spec", "after"}
					if len(log) != len(want) {
						t.Fatalf("log = %v, want %v", log, want)
					}
					for i := range want {
						if log[i] != want[i] {
							t.Fatalf("log = %v, want %v", log, want)
						}
					}
				})
			})
		})
	})
}

func TestRunAliased(t *testing.T) {
	spec.RunAliased(t, "RunAliased", func(t *testing.T, describe, context spec.Describe, it spec.S, before, after func(func())) {
		var log []string
		before(func() { log = append(log, "before") })
		after(func() { log = append(log, "after") })

		context("a nested group", func() {
			it("runs before, the spec, then after", func() {
				log = append(log, "spec")
				t.Cleanup(func() {
					want := []string{"before", "spec", "after"}
					if len(log) != len(want) {
						t.Fatalf("log = %v, want %v", log, want)
					}
					for i := range want {
						if log[i] != want[i] {
							t.Fatalf("log = %v, want %v", log, want)
						}
					}
				})
			})
		})
	})
}
