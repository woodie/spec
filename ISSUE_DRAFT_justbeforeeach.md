Ginkgo has [`JustBeforeEach`](https://onsi.github.io/ginkgo/#separating-creation-and-configuration-justbeforeeach):
a hook that runs after every `BeforeEach` at every nesting level, immediately
before the test body itself. It's what lets a suite separate "what varies"
(declared via ordinary `BeforeEach` in each nested container) from "the
action under test" (declared once, in a parent), instead of duplicating
that action in every leaf or hiding it behind a closure called explicitly
in each test.

`spec` has no equivalent today. The workaround (a `subject := func() { ... }`
closure over shared locals, called explicitly inside each `it`) gets to the
same place, but it's a workaround, not the real thing -- it has to be
invoked at each call site rather than running automatically, and every
value it closes over has to be a plain local rather than assignable
straight into a shared `var` the way `Before`/`After` already allow.

This isn't a hypothetical want. The Kotest ecosystem hit the identical gap
and discussed it at length: [kotest/kotest#952](https://github.com/kotest/kotest/issues/952),
filed in 2019 against KotlinTest (Kotest's predecessor), asked for exactly
this -- explicitly citing Ginkgo's `JustBeforeEach` as the model -- and sat
closed with no resolution. We ended up solving it for Kotest ourselves
with a small extension ([`kwick`](https://github.com/woodie/kwick)'s
`JustBeforeEachExtension`), which suggests this is a real, recurring gap
in BDD-style frameworks generally, not one specific to `spec`.

## Proposed shape

Looking at `spec.go`'s existing hook mechanism (`specHooks`/`specHook`,
the linked list built via `hooks.next()` as the tree is walked, run via
`specHooks.run`), the addition looks surgical rather than invasive:

- A new `S.JustBefore(f func())` method, declared the same way `Before`/
  `After` are:
  ```go
  func (s S) JustBefore(f func()) {
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

That's the shape as best we can tell from reading the source -- happy to
turn this into a real PR (with its own spec coverage) if that's useful,
rather than just leaving it as a request.
