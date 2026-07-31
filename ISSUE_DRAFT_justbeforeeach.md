**Breaking change, intentionally.** This fork is free to make a clean
break from upstream naming where it's worth it -- and cross-language
naming consistency is worth more now than it was when `spec` was written:
`Before`/`After` become `BeforeEach`/`AfterEach`, and a new
`JustBeforeEach` is added alongside them. All three names match what
Ginkgo, RSpec, Jest/Mocha, and Kotest already call the same three hooks,
so porting a suite (or a person's mental model) between languages stops
requiring a translation table.

## Rename: `Before`/`After` -> `BeforeEach`/`AfterEach`

`S.Before`/`S.After` (and `Suite.Before`/`Suite.After`) already run
"before/after each spec in the group/suite" per their own godoc comments
-- genuinely unambiguous "each" semantics, at every nesting level, with no
once-per-group equivalent anywhere in `spec` (deliberate, per the READMEs
own "does not reuse any closures between test runs" principle -- there's
no `BeforeAll`/`BeforeSuite`-style hook to confuse `Before` with). The
behavior's already right; the name just doesn't say so, and every other
BDD-style framework uses the "Each" suffix specifically to promise "reruns
every time, nothing shared." Renaming closes that gap outright instead of
leaving a same-meaning-different-name alias around indefinitely:

```go
// BeforeEach runs a function before each spec in the group.
func (s S) BeforeEach(f func()) {
    s("", f, func(c *config) { c.before = true })
}

// AfterEach runs a function after each spec in the group.
func (s S) AfterEach(f func()) {
    s("", f, func(c *config) { c.after = true })
}
```

(Same shape for `Suite.BeforeEach`/`Suite.AfterEach`.)

This is a real breaking change: every existing `it.Before(...)`/
`it.After(...)` call site stops compiling. That's deliberate -- a
compile-time break (the old method simply doesn't exist anymore) is
better than keeping `Before`/`After` around as deprecated shims that panic
at runtime, since it surfaces every call site immediately, at build time,
with the compiler pointing at the exact line, rather than waiting for that
specific test to execute. No migration shim, no permanent dead code kept
around just to fail loudly later.

## New: `JustBeforeEach`

Ginkgo has [`JustBeforeEach`](https://onsi.github.io/ginkgo/#separating-creation-and-configuration-justbeforeeach):
a hook that runs after every `BeforeEach` at every nesting level,
immediately before the test body itself. It's what lets a suite separate
"what varies" (declared via ordinary `BeforeEach` in each nested
container) from "the action under test" (declared once, in a parent),
instead of duplicating that action in every leaf or hiding it behind a
closure called explicitly in each test.

`spec` has no equivalent today. The workaround (a `subject := func() { ... }`
closure over shared locals, called explicitly inside each `it`) gets to the
same place, but it's a workaround, not the real thing -- it has to be
invoked at each call site rather than running automatically, and every
value it closes over has to be a plain local rather than assignable
straight into a shared `var` the way `BeforeEach`/`AfterEach` already allow.

This isn't a hypothetical want. The Kotest ecosystem hit the identical gap
and discussed it at length: [kotest/kotest#952](https://github.com/kotest/kotest/issues/952),
filed in 2019 against KotlinTest (Kotest's predecessor), asked for exactly
this -- explicitly citing Ginkgo's `JustBeforeEach` as the model -- and sat
closed with no resolution. We ended up solving it for Kotest ourselves
with a small extension ([`kwick`](https://github.com/woodie/kwick)'s
`JustBeforeEachExtension`), which suggests this is a real, recurring gap
in BDD-style frameworks generally, not one specific to `spec`.

### Proposed shape

Looking at `spec.go`'s existing hook mechanism (`specHooks`/`specHook`,
the linked list built via `hooks.next()` as the tree is walked, run via
`specHooks.run`), the addition looks surgical rather than invasive:

- A new `S.JustBeforeEach(f func())` method, declared the same way
  `BeforeEach`/`AfterEach` are:
  ```go
  func (s S) JustBeforeEach(f func()) {
      s("", f, func(c *config) { c.justBefore = true })
  }
  ```
- A new `config.justBefore` flag, wired into `Run`'s inner switch
  alongside the existing `cfg.before`/`cfg.after` cases.
- `specHook` gains a `justBefore []func()` slice alongside its existing
  `before`/`after`.
- `specHooks.run` collects every level's `justBefore` funcs while walking
  `before`s (outermost to innermost, same order `before` itself already
  runs in), then runs all of them, in that same order, after the full
  `before` loop completes and immediately before `spec()`:
  ```go
  func (s specHooks) run(t *testing.T, spec func()) {
      t.Helper()
      var justBefores []func()
      for h := s.first; h != nil; h = h.next {
          defer run(t, h.after...)
          run(t, h.before...)
          justBefores = append(justBefores, h.justBefore...)
      }
      run(t, justBefores...)
      run(t, spec)
  }
  ```

## Fallout, named up front

Every consumer in this account that aliases `it.Before`/`it.After`
(`gorderly`, `expect`, `humane`, `lambada` -- the `context, before, after
:= describe, it.Before, it.After` line documented in `gorderly`'s own
`docs/FRAMEWORK.md`) breaks at compile time the moment its `go.mod` picks
up this version. Same shape as the `Expect(t, got)` -> `Expect(got, t)`
break `expect` shipped as `v0.2.0` -- a real, one-time ripple through every
consumer's test files, done deliberately in one pass rather than left
half-migrated. Not a reason to avoid the rename, just a real follow-up
task once this ships.

The convention itself changes too, not just the names in it. Aliasing all
three renamed/new methods to bare lowercase locals doesn't hold up --
`context, beforeEach, afterEach, justBeforeEach := describe, it.BeforeEach,
it.AfterEach, it.JustBeforeEach` is a genuinely long, cluttered line, and
abbreviating it (`it.BE`/`it.JBE`) just trades one kind of ugly for
another. Simpler fix: stop aliasing `it`'s hook methods at all. `it`
already reads visually distinct from the lowercase `describe`/`context`
structural vocabulary, so `it.BeforeEach(...)`/`it.AfterEach(...)`/
`it.JustBeforeEach(...)` called qualified doesn't clash the way a
qualified `expect.Expect(...)` would. The alias line shrinks back down to
just `context := describe`, and every consumer's call sites change from
`before(func() { ... })` to `it.BeforeEach(func() { ... })` (etc.) --
mechanical, but real, across every test file in every consumer, confirmed
on the user's own Mac per consumer, one at a time.

Happy to turn this into a real PR (with its own spec coverage for
`JustBeforeEach` and updated coverage for the renamed methods) if that's
useful, rather than just leaving it as a request.
