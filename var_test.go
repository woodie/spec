package spec_test

import (
	"testing"

	"github.com/sclevine/spec"
)

func TestVar(t *testing.T) {
	spec.Run(t, "Var", func(t *testing.T, when spec.G, it spec.S) {
		when("Set then Get", func() {
			v := spec.NewVar[int]("count")
			it.Before(func() { v.Set(5) })

			it("returns the stored value", func() {
				if v.Get() != 5 {
					t.Errorf("Get() = %d, want 5", v.Get())
				}
			})
		})

		when("Get before Set", func() {
			v := spec.NewVar[int]("count")

			it("panics with a clear message", func() {
				defer func() {
					if recover() == nil {
						t.Fatal("expected Get to panic before Set")
					}
				}()
				v.Get()
			})
		})

		when("re-evaluated fresh per spec, not once per when", func() {
			v := spec.NewVar[int]("count")
			it.Before(func() { v.Set(0) })

			it("mutates in the first it", func() {
				v.Set(v.Get() + 1)
				if v.Get() != 1 {
					t.Errorf("Get() = %d, want 1", v.Get())
				}
			})

			it("does not see the first it's mutation", func() {
				if v.Get() != 0 {
					t.Errorf("Get() = %d, want 0 -- Var did not reset fresh for this it", v.Get())
				}
			})
		})
	})
}
