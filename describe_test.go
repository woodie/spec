package spec_test

import (
	"testing"

	"github.com/sclevine/spec"
)

func TestAsContext(t *testing.T) {
	spec.Run(t, "AsContext", func(t *testing.T, describe spec.Describe, it spec.S) {
		context := describe.AsContext()

		context("runs specs exactly like describe", func() {
			it("gets here", func() {})
		})
	})
}
