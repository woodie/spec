# spec ⑂

[![go.mod version](https://img.shields.io/github/go-mod/go-version/woodie/spec)](https://github.com/woodie/spec)
[![CI](https://github.com/woodie/spec/actions/workflows/ci.yml/badge.svg)](https://github.com/woodie/spec/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/woodie/spec.svg)](https://github.com/woodie/spec/releases/latest)
[![License](https://img.shields.io/github/license/woodie/spec.svg)](LICENSE)

This is [`woodie`](https://github.com/woodie)'s fork of
[`sclevine/spec`](https://github.com/sclevine/spec) -- see
[upstream's own README](https://github.com/sclevine/spec/blob/master/README.md)
for the original pitch and history. This fork exists to move `spec`'s hook
naming into line with what Ginkgo, RSpec, Jest/Mocha, and Kotest already
call the same three hooks (`BeforeEach`/`AfterEach`/`JustBeforeEach`),
since cross-language naming consistency is worth more now than it was when
`spec` was written, and upstream looks maintained-but-dormant rather than
positioned to make that change itself.

**Status:** `it.BeforeEach`/`it.AfterEach`/`it.JustBeforeEach` are here.
`it.Before`/`it.After` still work too -- kept as deprecated aliases so no
existing consumer breaks -- but `golangci-lint`'s default `staticcheck`
check will flag them as deprecated. See `docs/COWORK.md` for the full
reasoning and rollout plan.

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

### Examples

[Most functionality is demonstrated here.](spec_test.go#L238)

Quick example:

```go
func TestObject(t *testing.T) {
    spec.Run(t, "object", func(t *testing.T, context spec.G, it spec.S) {
        describe := context

        var someObject *myapp.Object

        it.BeforeEach(func() {
            someObject = myapp.NewObject()
        })

        it.AfterEach(func() {
            someObject.Close()
        })

        it("should have some default", func() {
            if someObject.Default != "value" {
                t.Error("bad default")
            }
        })

        context("something happens", func() {
            it.BeforeEach(func() {
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

        describe("some slow things happen", func() {
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

### `JustBeforeEach`

```go
spec.Run(t, "Calculator", func(t *testing.T, context spec.G, it spec.S) {
    describe := context

    var calculator *Calculator
    it.BeforeEach(func() { calculator = NewCalculator() })

    describe("#divide", func() {
        var numerator, denominator, result int
        it.JustBeforeEach(func() {
            result = calculator.Divide(numerator, denominator)
        })

        context("dividing evenly", func() {
            it.BeforeEach(func() { numerator, denominator = 10, 2 })

            it("returns the quotient", func() {
                if result != 5 {
                    t.Error("expected 5")
                }
            })
        })

        context("dividing with a remainder", func() {
            it.BeforeEach(func() { numerator, denominator = 7, 2 })

            it("truncates toward zero", func() {
                if result != 3 {
                    t.Error("expected 3")
                }
            })
        })
    })
})
```

`it.JustBeforeEach` runs after every `it.BeforeEach` at every nesting
level, immediately before the test itself -- so `numerator`/`denominator`
are always set before `calculator.Divide` runs, and each `context` only
states what's different about its own inputs. Same shape as Ginkgo's
`JustBeforeEach`; see `docs/COWORK.md` for the full reasoning.

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
