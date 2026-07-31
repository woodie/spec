# Working with `spec` (the `woodie/spec` fork)

Cross-project conventions are in `~/workspace/woodie/docs/COWORK.md`.

## What this is

A personal fork of [`sclevine/spec`](https://github.com/sclevine/spec),
currently sitting on plain `master` with no divergence from upstream.
`expect`, `gorderly`, `humane`, and every other consumer in this account
depend on plain upstream `sclevine/spec` directly, not this fork -- see
"Reversal: moved off the `woodie/spec` fork" in `expect`'s/`gorderly`'s/
`humane`'s own `docs/COWORK.md` for why an earlier version of this fork
(built around a `RunAliased` convenience wrapper) got walked back. That
reversal was about unnecessary API surface, not a decision to never use
this fork again -- see below for why it's back in play.

## Upstream looks dormant

Emailed Stephen Levine (`sclevine`, upstream's author) directly on Jul 18,
cc'ing Caleb Bron, Joe Fitzgerald, and Zach Gershman (other Ginkgo/Gomega-
adjacent contributors), about `gorderly`'s spec-formatting work and
`expect`. No response after 13 days. Combined with upstream's own issue
tracker having issue creation restricted and a small fork/star count,
reasonable to treat upstream as maintained-but-not-actively-developed
rather than expect a reply.

Issues are now enabled on this fork
([`github.com/woodie/spec/issues`](https://github.com/woodie/spec/issues)),
so feature work can be filed and tracked here directly instead of waiting
on upstream.

## Planned: rename to `BeforeEach`/`AfterEach`, add `JustBeforeEach` -- breaking, deliberately

Decided directly: this fork will make a real breaking change, not just add
a feature. `Before`/`After` become `BeforeEach`/`AfterEach` (the old names
removed outright, not kept as deprecated shims), and a new
`JustBeforeEach` is added alongside them -- matching what Ginkgo, RSpec,
Jest/Mocha, and Kotest already call the same three hooks. Rationale,
stated directly: cross-language naming consistency is worth more now than
it was when `spec` was written seven years ago, and a compile-time break
(the old method just doesn't exist) is better than a runtime-panic
deprecation shim -- it surfaces every call site immediately at build time
instead of waiting for that specific test to run.

Why each piece:

- `S.Before`/`S.After` (and `Suite.Before`/`Suite.After`) already run
  "before/after each spec in the group/suite" per their own godoc
  comments -- genuinely unambiguous "each" semantics, at every nesting
  level, with no once-per-group equivalent anywhere in `spec` (deliberate,
  per the READMEs own "does not reuse any closures between test runs"
  principle). The behavior's already right; the name just didn't say so.
- Ginkgo has [`JustBeforeEach`](https://onsi.github.io/ginkgo/#separating-creation-and-configuration-justbeforeeach);
  `spec` doesn't. The account's own workaround is the `subject`-closure
  pattern (documented in `gorderly`'s `docs/FRAMEWORK.md`, "Why `spec`,
  not Ginkgo"/"The subject pattern") -- it gets to the same place but has
  to be invoked explicitly at each call site rather than running
  automatically. [kotest/kotest#952](https://github.com/kotest/kotest/issues/952)
  (filed 2019 against KotlinTest, closed with no resolution) shows the
  identical gap recurring in a completely different BDD ecosystem, citing
  Ginkgo's `JustBeforeEach` as the model too -- not a `spec`-specific
  complaint. We already solved this for Kotest with a small extension
  ([`kwick`](https://github.com/woodie/kwick)'s `JustBeforeEachExtension`).

Drafted a real issue (`ISSUE_DRAFT_justbeforeeach.md` in this repo's root,
not yet filed) covering both the rename and the new hook together, with a
concrete implementation shape sketched against `spec.go`'s actual hook
mechanism (the `specHooks`/`specHook` linked list, built via
`hooks.next()`, run via `specHooks.run`). Not yet filed, not yet
implemented, not yet confirmed against a real Go toolchain (no Go in this
sandbox -- see `~/workspace/woodie/docs/COWORK.md`'s "Working on
unfamiliar stacks").

**Fallout, named up front, not a surprise for later:** every consumer in
this account that aliases `it.Before`/`it.After` (`gorderly`, `expect`,
`humane`, `lambada` -- the `context, before, after := describe, it.Before,
it.After` line documented in `gorderly`'s own `docs/FRAMEWORK.md`) breaks
at compile time the moment its `go.mod` picks up this version. Same shape
as the `Expect(t, got)` -> `Expect(got, t)` break `expect` shipped as
`v0.2.0` -- a real, one-time ripple through every consumer's test files,
done deliberately in one pass rather than left half-migrated.

**The aliasing convention itself changes too.** Worked through this
directly: aliasing all three renamed/new methods to bare lowercase locals
(`context, beforeEach, afterEach, justBeforeEach := describe,
it.BeforeEach, it.AfterEach, it.JustBeforeEach`) is a genuinely long,
cluttered line, and abbreviating it (`it.BE`/`it.JBE`) just trades one
kind of ugly for another. Landed on the simpler fix instead: stop aliasing
`it`'s hook methods at all. `it` already reads visually distinct from the
lowercase `describe`/`context` structural vocabulary, so calling
`it.BeforeEach(...)`/`it.AfterEach(...)`/`it.JustBeforeEach(...)` qualified
doesn't clash the way a qualified `expect.Expect(...)` would (the reason
`expect` itself is dot-imported). The alias line shrinks back down to just
`context := describe`; every consumer's call sites move from
`before(func() { ... })` to `it.BeforeEach(func() { ... })` (etc.).

## Done: README refreshed around this fork's own positioning

Rewrote the opening to point explicitly at
[upstream's own README](https://github.com/sclevine/spec/blob/master/README.md)
and state directly why this fork exists (naming consistency, upstream
dormancy) and its current status (rename/`JustBeforeEach` planned, not
shipped -- this checkout still behaves exactly like upstream today).
Dropped the Travis/GoDoc badges -- both pointed at upstream's own
`sclevine/spec` CI/pkg.go.dev entries, not this fork's, so they were
misleading rather than just stale. No CI configured on this fork yet
(only the old `.travis.yml`, and Travis itself is effectively dead for
projects like this); worth a real GitHub Actions workflow (matching
`gorderly`'s/`expect`'s own `ci.yml` shape) once there's real code to test
against, not before.

## Done: README example renamed `when` -> `describe`/`context`

Separate from the rename above, and no code change at all --
`spec.G`/`spec.S` are just function types, upstream never tied any
naming to them beyond the example. Upstream's own README named the group
parameter `when` (`func(t *testing.T, when spec.G, it spec.S)`), which
never matched this account's actual convention (`describe` at the top
level, `context := describe` aliased for nested groups, documented in
every consumer's own `docs/FRAMEWORK.md`). Updated the fork's README
example directly to `describe`/`context` so the fork's own docs describe
how this account actually writes `spec` suites, not upstream's original
naming. Left `it.Before`/`it.After` as-is in that same example for now,
since the rename above hasn't shipped yet -- update this example to
`it.BeforeEach`/`it.AfterEach` in the same pass that implements the
rename, so the README never describes a method that doesn't exist yet.
`spec_test.go`/`options_test.go` (upstream's own test suite) still use
`when` -- left untouched for now, since rewriting real test code is a
bigger, riskier change than a README example; worth a separate pass if
wanted.

Next steps once the issue's filed: implement on a real branch (rename
plus the new hook), add/update spec coverage, confirm on the user's own
Mac, then work through each consumer one at a time (`gorderly`, `expect`,
`humane`, `lambada`) updating both the aliasing line and every call site,
confirming clean per consumer -- same "publish first, then bump each
consumer's pin" sequence as any other shared-library change, per the
cross-project doc's "Shared libraries across sibling repos" section.
`gorderly`'s own `docs/FRAMEWORK.md` will also need its "Why `spec`, not
Ginkgo"/"subject pattern"/"Aliasing `spec`'s structural functions"
sections updated once this ships, since they currently describe the gap
this fork is about to close and the aliasing convention this fork is
about to change.
