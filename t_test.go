package spec_test

import (
	"os"
	"testing"

	"github.com/sclevine/spec"
)

func TestT(t *testing.T) {
	spec.Run(t, "T", func(t *testing.T, when spec.G, it spec.S) {
		when("TempDir is called from Before, mirroring GinkgoT().TempDir()", func() {
			var dir string
			it.Before(func() { dir = it.T().TempDir() })

			it("returns a real, existing directory", func() {
				info, err := os.Stat(dir)
				if err != nil {
					t.Fatalf("TempDir() dir doesn't exist: %v", err)
				}
				if !info.IsDir() {
					t.Fatalf("TempDir() returned a non-directory: %s", dir)
				}
			})
		})
	})
}
