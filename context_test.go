package spec_test

import (
	"context"
	"testing"

	"github.com/sclevine/spec"
)

func TestContext(t *testing.T) {
	spec.Run(t, "Context", func(t *testing.T, when spec.G, it spec.S) {
		it("returns a non-nil, not-yet-canceled context", func() {
			ctx := it.Context()
			if ctx == nil {
				t.Fatal("Context() returned nil")
			}
			if ctx.Err() != nil {
				t.Fatal("context canceled before spec finished:", ctx.Err())
			}
		})

		when("Context is read from Before", func() {
			var ctx context.Context
			it.Before(func() { ctx = it.Context() })

			it("captured a non-nil context", func() {
				if ctx == nil {
					t.Fatal("Context() returned nil when called from Before")
				}
			})
		})

		it("cancels the context before Cleanup runs", func() {
			ctx := it.Context()
			t.Cleanup(func() {
				if ctx.Err() == nil {
					t.Error("expected context to be canceled by Cleanup time")
				}
			})
		})
	})
}
