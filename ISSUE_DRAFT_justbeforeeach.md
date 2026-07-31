**Add `BeforeEach`/`AfterEach`/`JustBeforeEach`, deprecate `Before`/`After`.**
Cross-language naming consistency is worth more now than it was when
`spec` was written: `Before`/`After` gain `BeforeEach`/`AfterEach` as their
real names, and a new `JustBeforeEach` joins them. All three match what
Ginkgo, RSpec, Jest/Mocha, and Kotest already call the same three hooks,
so porting a suite (or a person's mental model) between languages stops
requiring a translation table.

## Rename: `Before`/`After` -> `BeforeEach`/`AfterEach` (old names kept, deprecated)

`S.Before`/`S.After` (and `Suite.Before`/`Suite.After`) already run
"before/after each spec in the group/suite" per their own godoc comments
-- genuinely unambiguous "each" semantics, at every nesting level, with no
once-per-group equivalent anywhere in `spec` (deliberate, per the READMEs
own "does not reuse any closures between test runs" principle -- there's
no `BeforeAll`/`BeforeSuite`-style hook to confuse `Before` with). The
behavior's already right; the name just doesn't say so, and every other
BDD-style framework uses the "Each" suffix specifically to promise "reruns
every time, nothing shared."

Rather than remove `Before`/`After` outright, they stay as thin deprecated
aliases:

```go
// BeforeEach runs a function before each spec in the group.
func (s S) BeforeEach(f func()) {
    s("", f, func(c *config) { c.before = true })
}

// AfterEach runs a function after each spec in the group.
func (s S) AfterEach(f func()) {
    s("", f, func(c *config) { c.after = true })
}

// Before runs a function before each spec in the group.
//
// Deprecated: use BeforeEach instead. Before is kept as an alias so
// existing call sites keep compiling; it will be removed in a future
// major version.
func (s S) Before(f func()) {
    s.BeforeEach(f)
}

// After runs a function after each spec in the group.
//
// Deprecated: use AfterEach instead. After is kept as an alias so
// existing call sites keep compiling; it will be removed in a future
// major version.
func (s S) After(f func()) {
    s.AfterEach(f)
}
```

(Same shape for `Suite.BeforeEach`/`Suite.AfterEach`, with `Suite.Before`/
`Suite.After` kept as the same style of deprecated alias.)

No consumer's build breaks the moment it picks up this version. The
standard `// Deprecated:` godoc comment is enough to surface the old names
as a real warning without a runtime shim or a compile break: gopls/most
IDEs render deprecated symbols with a strikethrough, and `staticcheck`'s
`SA1019` check -- part of `golangci-lint`'s default linter set, so nothing
extra to configure -- fails `make check`/`make lint` wherever `it.Before`/
`it.After` are still called. That gets every call site the same "you need
to look at this" signal a compile error would, just via lint instead of
the compiler, and leaves each consumer free to migrate on its own
schedule rather than all at once. Chosen over a hard break specifically
because a real break would make this fork painful to adopt for anyone not
already planning a coordinated migration -- including, potentially,
upstream itself, if `sclevine` ever wants to pull this back in.

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
`docs/FRAMEWORK.md`) keeps compiling and passing `go test` the moment its
`go.mod` picks up this version -- `Before`/`After` are still there, just
deprecated. What changes is `make check`/`make lint`: `staticcheck`'s
SA1019 starts flagging every `it.Before(...)`/`it.After(...)` call site as
deprecated-symbol usage, the same way it already flags any other
deprecated stdlib or third-party API. Unlike the `Expect(t, got)` ->
`Expect(got, t)` break `expect` shipped as `v0.2.0` (a real compile
break), this is a lint signal each consumer can act on at its own pace --
still a real follow-up task once this ships, just not a blocking one.

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
