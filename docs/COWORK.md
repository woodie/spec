# Working with spec

Cross-project conventions are in `~/workspace/woodie/docs/COWORK.md`.

## What this is

A fork of `sclevine/spec` (last upstream tag `v1.4.0`, December 2019),
kept at the same module path (`github.com/sclevine/spec`) so any consumer
can drop it in via a `replace` directive with zero import-path changes.
Adopted for `gorderly`'s own tests and `lambada`'s migration off
Ginkgo/Gomega (see both repos' own `docs/COWORK.md`) once real usage there
showed what a modern-Go-flavored version of `spec` could add without
touching its actual execution model.

## Why fork instead of dropping `spec` for something else

Evaluated directly against hand-rolling `describe`/`context`/`it` on bare
`t.Run` (see `gorderly`'s "Language-mechanism exploration" and
"Test-writing convention" sections). The reason `spec` won isn't
staleness-related and nothing Go has shipped since 2019 -- generics,
`t.Context()`, loop-var semantics, range-over-func -- closes the actual gap
`spec` fills: automatic re-run of `Before`/`After` per leaf spec, matching
Kotest/Quick's automatic `beforeEach`, which a hand-rolled version can't get
without the developer manually re-invoking setup inside every `it`. `spec`
also already routes every leaf through a real `t.Run(name, func(t
*testing.T){...})` call (verified by reading `parser.go`, not the README),
so `go test -v` output stays completely standard -- unlike Ginkgo, which
owns its own execution/reporting engine and needed `ginkgo-fd`'s
`--json-report` round-trip to get data back out at all.

## Additions this fork carries over upstream

Each is additive (nothing upstream removed or changed in place) and scoped
to be pulled out as its own PR if `sclevine/spec`'s owner ever wants any of
it back -- listed separately on purpose, not as one lump diff.

- **`S.Context()`** -- returns the spec's `context.Context` (nil outside
  `S`), canceled when the spec completes. Same side-channel pattern `Out()`
  already used (`c.ctx func(context.Context)` in `config`, set via an
  `Option` that `Run`'s per-spec switch resolves with `t.Context()`).
  Excluded from spec-counting in `parser.go`'s parse phase the same way
  `Out()` already was -- a real, necessary fix, not just symmetry (missing
  it double-counts a phantom spec node).
- **`S.T()`** -- returns the real `*testing.T` (nil outside `S`), same
  side-channel shape. Justified by `lambada`'s actual Ginkgo usage:
  `GinkgoT().TempDir()` inside `BeforeEach` and `Skip(...)` inside
  `BeforeEach` both need the real `*testing.T`, not just its
  `context.Context` -- `Before`/`After` take `func()` with no receiver, so
  without `T()` there was no way to reach either from inside a hook at all.
- **`Describe` (`type Describe = G`)** -- pure naming sugar, zero behavior
  change, so a suite's outermost group can read `describe` instead of the
  generic `G`, matching this account's Quick/Kotest `describe`/`context`
  convention.
- **`G.AsContext()`** (`func (g G) AsContext() G { return g }`) -- also a
  no-op, names the `context := describe.AsContext()` alias idiom instead of
  a bare, easy-to-miss assignment. Deliberately not a second type alias
  (`type Context = G`) -- that name was already taken by `S.Context()`
  above, and reusing it for a second, unrelated meaning in the same small
  package was judged more confusing than a distinct method name.
- **`Var[T]`** (`var.go`) -- generics-based `NewVar[T](name)`/`Set`/`Get`,
  for the "set up context above, plug into it below" pattern (see
  `next-caltrain-swift`/`next-caltrain-kotlin`'s `GoodTimesSpec` files).
  Panics with a clear message on `Get` before `Set` in a given spec, instead
  of silently returning a zero value -- everything else about per-spec
  freshness is already free from `spec`'s own re-evaluation model, so `Var`
  adds nothing there.
- **`Aliases(describe, it) (before, after func(func()), context G)`** --
  bundles `it.Before`, `it.After`, `describe.AsContext()` into one call,
  replacing what had been creeping up to a 3-part multiple-assignment line
  per file.
- **`RunAliased(t, text, f, opts...)`** -- wraps `Run`, passing
  `describe`/`context`/`it`/`before`/`after` straight into `f`'s own
  parameters via `Aliases`. The actual answer to "make this work always
  without typing it" -- `Aliases` alone still needed one line per file;
  `RunAliased` needs none, at any nesting depth, since Go closures already
  see every enclosing scope's locals. (First attempt at this put the
  wrapper in `lambada`'s own `internal/spectest` package, reasoning it was
  project-specific like a Ruby `spec_helper.rb` -- wrong call, corrected
  same session: there's nothing `lambada`-specific about it, so it belongs
  here, where every consumer of this fork gets it.)

## Verification

No Go toolchain in this sandbox -- every change above was written by
inspection, matching `gorderly`/`lambada`'s own documented sandbox
limitation. On your Mac:

```
cd ~/workspace/spec
go test -v ./...
```
