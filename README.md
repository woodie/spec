# spec

[![Build Status](https://travis-ci.org/sclevine/spec.svg?branch=master)](https://travis-ci.org/sclevine/spec)
[![GoDoc](https://godoc.org/github.com/sclevine/spec?status.svg)](https://godoc.org/github.com/sclevine/spec)

This is a fork of [`sclevine/spec`](https://github.com/sclevine/spec) --
a simple BDD test organizer that runs on real `go test` subtests, no
framework of its own, no assertions, no global state. For the full case
for the base library (its design, its non-goals, its usage examples),
**see [the original README](https://github.com/sclevine/spec#readme)** --
this file only covers what this fork changes and why.

Kept at the same import path (`github.com/woodie/spec`) so it can be
pulled in via a `go.mod` `replace` directive with zero call-site changes if
you're already using upstream `spec`.

## Why fork instead of just using upstream

`sclevine/spec`'s last tag is from 2019. Its core design still holds up --
real subtests, no test pollution between leaves, no assertion library
opinion -- but Go has grown two features since then that let it do a bit
more without giving up any of that: generics (1.18) and `t.Context()`
(1.24). This fork is that gap, closed, as a small set of additions rather
than a rewrite.

## What this fork adds

- **`it.Context()`** -- passes through the real `t.Context()` for the
  running subtest, so specs can hand a `context.Context` to code under
  test.
- **`it.T()`** -- passes through the real `*testing.T` for the running
  subtest, so `Before`/`After` bodies can call `t.TempDir()`, `t.Skip()`,
  etc. without spec needing assertion or lifecycle features of its own.
- **`Describe`/`AsContext()`** -- `Describe` is a plain type alias for `G`;
  `describe.AsContext()` is a no-op method returning the same `G` under a
  second name, for RSpec-style `describe`/`context` duality without a
  second, colliding type alias (`Context` was already spoken for by
  `it.Context()` above).
- **`Var[T]`** -- a generics-based typed box (`NewVar[T]`, `.Set`, `.Get`)
  for sharing a value between a `Before` and its sibling `it` blocks
  without an `interface{}` cast at every read.
- **`Aliases`/`RunAliased`** -- `Aliases(describe, it)` returns `before,
  after, context` bound to `it.Before`, `it.After`, and
  `describe.AsContext()`; `RunAliased` wraps `Run` and passes all five
  (`describe`, `context`, `it`, `before`, `after`) directly as callback
  parameters, so no per-file alias line is needed anywhere. Works
  everywhere spec is imported -- no project-specific setup file required
  (Ruby's equivalent would be `spec_helper.rb`; Go's lack of implicit
  scoping means the closest analog is a small wrapper function, not a
  require).

Each addition is scoped to be independently upstreamable as its own PR;
see `docs/COWORK.md` for the reasoning behind each one and the existing
spec mechanism (the same side-channel `Option` pattern upstream's own
`Out()` already uses) each is built on.

A companion library, [`github.com/woodie/expect`](https://github.com/woodie/expect),
provides Gomega-style matchers (`Expect(t, x).To(Equal(y))`) for specs
written with this fork, without adopting Gomega itself.

## Using the additions together

```go
func TestObject(t *testing.T) {
    spec.RunAliased(t, "object", func(t *testing.T, describe, context spec.Describe, it spec.S, before, after func(func())) {
        var obj *myapp.Object

        before(func() {
            obj = myapp.NewObject(it.Context())
        })

        after(func() {
            obj.Close()
        })

        describe("something happens", func() {
            context("with a temp dir", func() {
                before(func() {
                    obj.Dir = it.T().TempDir()
                })

                it("does the thing", func() {
                    if err := obj.DoThing(); err != nil {
                        t.Error(err)
                    }
                })
            })
        })
    })
}
```

Everything else -- `Focus`/`Pend`, `Random`/`Reverse`/`Parallel`, `Report`,
multi-file `Suite`s -- is unchanged from upstream; see [its
README](https://github.com/sclevine/spec#readme) for those.
