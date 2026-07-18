# spec

[![Build Status](https://travis-ci.org/sclevine/spec.svg?branch=master)](https://travis-ci.org/sclevine/spec)
[![GoDoc](https://godoc.org/github.com/sclevine/spec?status.svg)](https://godoc.org/github.com/sclevine/spec)

Spec is a simple BDD test organizer for Go. It minimally extends the standard
library `testing` package by facilitating easy organization of Go 1.7+
[subtests](https://blog.golang.org/subtests).

Spec differs from other BDD libraries for Go in that it:
- Does not reimplement or replace any functionality of the `testing` package
- Does not provide an alternative test parallelization strategy to the `testing` package
- Does not provide assertions
- Does not encourage the use of dot-imports
- Does not reuse any closures between test runs (to avoid test pollution)
- Does not use global state, excessive interface types, or reflection

Spec is intended for gophers who want to write BDD tests in idiomatic Go using
the standard library `testing` package. Spec aims to do "one thing right,"
and does not provide a wide DSL or any functionality outside of test
organization.

### This fork: extending spec for the current state of Go

This fork (`github.com/woodie/spec`, drop-in at the same import path so it
can be pulled in via a `replace` directive with no call-site changes) adds a
small set of features that Go's evolution since spec's original release now
makes possible, without changing spec's own no-global-state, no-assertions,
no-reimplemented-`testing` philosophy:

- **`it.Context()`** -- passes through the real `t.Context()` for the running
  subtest, so specs can pass a `context.Context` to code under test.
- **`it.T()`** -- passes through the real `*testing.T` for the running
  subtest, so `Before`/`After` bodies can call `t.TempDir()`, `t.Skip()`, etc.
  without spec needing to grow assertion or lifecycle features of its own.
- **`Describe`/`AsContext()`** -- `Describe` is a plain type alias for `G`;
  `describe.AsContext()` is a no-op method that returns the same `G` under a
  second name, for RSpec-style `describe`/`context` duality without a second
  colliding type alias (`Context` was already spoken for by `it.Context()`
  above).
- **`Var[T]`** -- a generics-based typed box (`NewVar[T]`, `.Set`, `.Get`) for
  sharing a value between a `Before` and its sibling `it` blocks without an
  `interface{}` cast at every read.
- **`Aliases`/`RunAliased`** -- `Aliases(describe, it)` returns
  `before, after, context` bound to `it.Before`, `it.After`, and
  `describe.AsContext()`; `RunAliased` wraps `Run` and passes all five
  (`describe`, `context`, `it`, `before`, `after`) directly as callback
  parameters, so no per-file alias line is needed. This works everywhere spec
  is imported -- it does not require any project-specific setup file (the Ruby
  equivalent would be `spec_helper.rb`, but Go's lack of implicit/global
  scoping means the closest analog is a small wrapper function, not a require).

None of the above were possible without generics (`Var[T]`) or without
`t.Context()` existing on `*testing.T` (`it.Context()`) -- both landed after
spec's original design. Each is scoped as an independently upstreamable
addition; see `docs/COWORK.md` for the full reasoning behind each one and
which existing spec mechanism (the same side-channel `Option` pattern
`Out()` already used) each is built on.

A companion library, [`github.com/woodie/expect`](https://github.com/woodie/expect),
provides Gomega-style matchers (`Expect(t, x).To(Equal(y))`) for specs
written with this fork, without adopting Gomega itself.

### Features

- Clean, simple syntax
- Supports focusing and pending tests
- Supports sequential, random, reverse, and parallel test order
- Provides granular control over test order and subtest nesting
- Provides a test writer to manage test output
- Provides a generic, asynchronous reporting interface
- Provides multiple reporter implementations

### Notes

- Use `go test -v` to see individual subtests.

### Examples

[Most functionality is demonstrated here.](spec_test.go#L238)

Quick example:

```go
func TestObject(t *testing.T) {
    spec.Run(t, "object", func(t *testing.T, when spec.G, it spec.S) {
        var someObject *myapp.Object

        it.Before(func() {
            someObject = myapp.NewObject()
        })

        it.After(func() {
            someObject.Close()
        })

        it("should have some default", func() {
            if someObject.Default != "value" {
                t.Error("bad default")
            }
        })

        when("something happens", func() {
            it.Before(func() {
                someObject.Connect()
            })

            it("should do one thing", func() {
                if err := someObject.DoThing(); err != nil {
                    t.Error(err)
                }
            })

            it("should do another thing", func() {
                if result := someObject.DoOtherThing(); result != "good result" {
                    t.Error("bad result")
                }
            })
        }, spec.Random())

        when("some slow things happen", func() {
            it("should do one thing in parallel", func() {
                if result := someObject.DoSlowThing(); result != "good result" {
                    t.Error("bad result")
                }
            })

            it("should do another thing in parallel", func() {
                if result := someObject.DoOtherSlowThing(); result != "good result" {
                    t.Error("bad result")
                }
            })
        }, spec.Parallel())
    }, spec.Report(report.Terminal{}))
}
```

With less nesting:

```go
func TestObject(t *testing.T) {
    spec.Run(t, "object", testObject, spec.Report(report.Terminal{}))
}

func testObject(t *testing.T, when spec.G, it spec.S) {
    ...
}
```

For focusing/reporting across multiple files in a package:

```go
var suite spec.Suite

func init() {
    suite = spec.New("my suite", spec.Report(report.Terminal{}))
    suite("object", testObject)
    suite("other object", testOtherObject)
}

func TestObjects(t *testing.T) {
	suite.Run(t)
}

func testObject(t *testing.T, when spec.G, it spec.S) {
	...
}

func testOtherObject(t *testing.T, when spec.G, it spec.S) {
	...
}
```