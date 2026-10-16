# spec ⑂

[![go.mod version](https://img.shields.io/github/go-mod/go-version/woodie/spec)](https://github.com/woodie/spec)
[![CI](https://github.com/woodie/spec/actions/workflows/ci.yml/badge.svg)](https://github.com/woodie/spec/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/woodie/spec.svg)](https://github.com/woodie/spec/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/woodie/spec.svg)](https://pkg.go.dev/github.com/woodie/spec)
[![License](https://img.shields.io/github/license/woodie/spec.svg)](LICENSE)

This is [`woodie`](https://github.com/woodie)'s fork of
[`sclevine/spec`](https://github.com/sclevine/spec). This fork exists to
move `spec`'s hook naming into line with what Ginkgo, RSpec, Jest/Mocha, and
Kotest already call the same three hooks
(`BeforeEach`/`AfterEach`/`JustBeforeEach`), since cross-language naming
consistency is worth more now than it was when `spec` was written, and
upstream looks maintained-but-dormant rather than positioned to make that
change itself.

**Status:** `it.BeforeEach`/`it.AfterEach`/`it.JustBeforeEach` are here.
`it.Before`/`it.After` still work too -- kept as deprecated aliases so no
existing consumer breaks -- but `golangci-lint`'s default `staticcheck`
check will flag them as deprecated.

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

### Example

[Most functionality is demonstrated here.](spec_test.go#L238)

```go
func TestCalculator(t *testing.T) {
    spec.Run(t, "Calculator", func(t *testing.T, context spec.G, it spec.S) {
        describe := context

        var calculator *Calculator
        it.BeforeEach(func() { calculator = NewCalculator() })
        it.AfterEach(func() { calculator = nil })

        describe("#divide", func() {
            var numerator, denominator, result int
            it.BeforeEach(func() { numerator, denominator = 10, 2 })
            it.JustBeforeEach(func() {
                result = calculator.Divide(numerator, denominator)
            })

            it("returns the quotient", func() {
                if result != 5 {
                    t.Error("expected 5")
                }
            })

            context("when the numerator leaves a remainder", func() {
                it.BeforeEach(func() { numerator = 7 })

                it("drops the remainder", func() {
                    if result != 3 {
                        t.Error("expected 3")
                    }
                })
            })

            context("when the denominator is 1", func() {
                it.BeforeEach(func() { denominator = 1 })

                it("returns the numerator", func() {
                    if result != 10 {
                        t.Error("expected 10")
                    }
                })
            })
        })
    }, spec.Report(report.Terminal{}))
}
```

`#divide` sets default inputs in an ordinary `it.BeforeEach`, and each
`context` overrides only the one it's about. The call itself can't go in
that `it.BeforeEach`, because `spec` runs a parent's `it.BeforeEach` before
its children's: it would divide 10 by 2 every time, before any `context`
had changed an input. `it.JustBeforeEach` runs after every `it.BeforeEach`
at every nesting level, immediately before the test itself, so it always
sees the final inputs, and each `it` just checks `result`. Same shape as
Ginkgo's `JustBeforeEach`. In RSpec terms, the defaults play the part of
`let`, and `it.JustBeforeEach` plays `subject`, which RSpec can leave lazy
instead.

With less nesting:

```go
func TestObject(t *testing.T) {
    spec.Run(t, "object", testObject, spec.Report(report.Terminal{}))
}

func testObject(t *testing.T, context spec.G, it spec.S) {
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

func testObject(t *testing.T, context spec.G, it spec.S) {
	...
}

func testOtherObject(t *testing.T, context spec.G, it spec.S) {
	...
}
```

## Development

```
make build    # go build ./... -- spec is a library, nothing to install
make test     # verbose, dogfoods spec.Run against spec's own suite, piped through gorderly
make lint     # golangci-lint
make check    # terse: silent on success, full log on failure
```

`make test` pipes through [`gorderly`](https://github.com/woodie/gorderly)
for RSpec-style documentation output; without it installed, run
`go test -v ./...` or `go test ./...` directly instead.

## Learn more

- [`expect`](https://github.com/woodie/expect) -- the matcher half of this
  pairing; `spec` has no dependency on it.
- [`gorderly`'s docs/FRAMEWORK.md](https://github.com/woodie/gorderly/blob/main/docs/FRAMEWORK.md) --
  full suites combining `spec` + `expect`: context nesting, the `subject`
  pattern, stubbing, `httptest`, and interface test doubles.
