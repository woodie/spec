# Working with `spec` (the `woodie/spec` fork)

Cross-project conventions are in `~/workspace/woodie/docs/COWORK.md`.

## What this is

A fork of [`sclevine/spec`](https://github.com/sclevine/spec), published
under its own module path, `github.com/woodie/spec` (`v0.3.0`). It adds
`BeforeEach`/`AfterEach`/`JustBeforeEach` on top of upstream. `expect`,
`gorderly`, `humane`, and `lambada` all require it directly.

An earlier version of this fork, built around a `RunAliased` convenience
wrapper, was walked back (see "Reversal: moved off the `woodie/spec` fork"
in `expect`'s/`gorderly`'s/`humane`'s own `docs/COWORK.md`). That was about
unnecessary API surface, not a decision against the fork itself.

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

## Implemented: `BeforeEach`/`AfterEach`/`JustBeforeEach`, with `Before`/`After` kept as deprecated aliases

Original plan (recorded below, then reversed mid-implementation): a real
breaking change, `Before`/`After` removed outright in favor of
`BeforeEach`/`AfterEach`, plus a new `JustBeforeEach` -- matching what
Ginkgo, RSpec, Jest/Mocha, and Kotest already call the same three hooks.

Reversed directly, mid-implementation, once the actual fallout became
concrete: a hard break makes this fork painful to adopt for anyone not
already planning a coordinated migration across every consumer at once --
including, potentially, upstream itself, if `sclevine` ever wants to pull
this back in. Landed on the softer version instead: `BeforeEach`/
`AfterEach`/`JustBeforeEach` are the real, documented names now, but
`Before`/`After` (on both `S` and `Suite`) stay as thin aliases that just
call the `Each`-suffixed versions, marked with the standard Go
`// Deprecated:` godoc comment. This doesn't just sit there silently --
`staticcheck`'s `SA1019` check (part of `golangci-lint`'s default linter
set already used by this repo's own `make lint`/`make check`, nothing
extra to configure) fails on any `it.Before(...)`/`it.After(...)` call
site, and gopls/most IDEs render the deprecated symbol with a
strikethrough. That gets every consumer the same "you need to look at
this" signal a compile error would, just via lint instead of the
compiler, and lets each one migrate on its own schedule instead of all at
once in lockstep with this fork's release.

Rationale for the original (now-reversed) breaking-change framing, kept
here since the underlying naming argument didn't change, only the
migration mechanism:

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

Filed as [woodie/spec#1](https://github.com/woodie/spec/issues/1), covering
both the rename and the new hook, with the implementation sketched against
`spec.go`'s own hook mechanism (the `specHooks`/`specHook` linked list,
built via `hooks.next()`, run via `specHooks.run`). Shipped through PR #2
from the `beforeeach-justbeforeeach` branch and released as `v0.2.0`; the
issue is closed.

**Fallout, named up front, not a surprise for later:** every consumer in
this account that aliases `it.Before`/`it.After` (`gorderly`, `expect`,
`humane`, `lambada` -- the `context, before, after := describe, it.Before,
it.After` line documented in `gorderly`'s own `docs/FRAMEWORK.md`) keeps
compiling and passing `go test` the moment its `go.mod` picks up this
version -- `Before`/`After` are deprecated, not removed. What actually
changes for those consumers is `make check`/`make lint`: `staticcheck`'s
SA1019 (part of `golangci-lint`'s default set already, no config needed)
starts flagging every `it.Before(...)`/`it.After(...)` call site as
deprecated-symbol usage. Unlike the `Expect(t, got)` -> `Expect(got, t)`
break `expect` shipped as `v0.2.0` (a real compile break, forced all at
once), this is a lint signal each consumer can act on at its own pace --
still real follow-up work once this ships, just not a blocking one.

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

## Done: implemented on a feature branch, not `master`

Implementation work (the actual `spec.go`/`options.go`/`parser.go`
changes) happens on a `beforeeach-justbeforeeach` branch, merged back to
`master` via a pull request rather than committed directly -- decided
directly once real code was in flight, since nothing else depends on
`master` yet and a PR gives a normal review/CI-gate point before this
lands, same as any other repo in this account. `gh pr create` needs to run
from the user's own Mac (no `gh`/network access in this sandbox), same
handoff pattern as `gh issue create` for the drafted issue above.

## Done: real CI, badges, dropped Travis

Added `.github/workflows/ci.yml` (checkout, setup-go pinned to `1.24` --
matching `go.mod`'s directive -- build, test) on push/PR to `master`
(this fork's actual default branch, not `main`), matching `expect`'s/
`gorderly`'s own `ci.yml` shape. Removed `.travis.yml` -- Travis CI is
effectively dead for a repo like this, and keeping it around next to a
real, working GitHub Actions workflow just invites confusion about which
one is authoritative.

README badges added: go.mod version, CI, Release (`v0.1.0` is already
tagged on `woodie/spec` from the earlier abandoned fork attempt), License.
Deliberately did **not** add a Go Reference/pkg.go.dev badge -- `go.mod`
still declares `module github.com/sclevine/spec` (unchanged from
upstream), so pkg.go.dev indexes this fork's code under upstream's own
module path, not `github.com/woodie/spec`. A badge pointing at either path
would be misleading: `github.com/woodie/spec` has nothing published to
pkg.go.dev under that path, and `github.com/sclevine/spec` shows
upstream's docs, not this fork's actual (soon-to-be-renamed) API. Whether
to ever rename the module path itself to `github.com/woodie/spec` is a
real open question -- bigger than this session's scope, and orthogonal to
the `Before`/`After` rename -- flagged here rather than decided.

## Done: first `make check` run, real lint fixes

First `make check` on the user's own Mac surfaced 5 real `golangci-lint`
issues, none related to the Makefile itself -- pre-existing debt the repo
never had lint coverage to catch before:

- Two `errcheck` hits in `options_test.go`: `fmt.Fprint(it.Out(), ...)`
  calls with unchecked return values. Fixed with `_, _ = fmt.Fprint(...)`,
  matching the established convention elsewhere in this account (e.g.
  `gorderly`'s `v0.3.0`/`v0.3.1` errcheck fix).
- Three `govet` "inline" warnings: `ioutil.ReadAll` (declared using
  go1.26.2, the installed toolchain) couldn't be inlined into files
  declaring `go 1.13` (`options_test.go`, `report/log.go`,
  `report/terminal.go`) -- a real gap between this module's 2019-era `go`
  directive and every sibling repo's (`expect`/`gorderly`/`humane` all
  sit at `go 1.24`/`1.25.0`). Fixed at the root rather than papered over:
  replaced every `ioutil.ReadAll` call with `io.ReadAll` (the modern,
  non-deprecated equivalent -- `io/ioutil`'s version is just a thin
  wrapper around it and has been deprecated since Go 1.16), dropping the
  `io/ioutil` import entirely. Also bumped `go.mod`'s `go` directive from
  `1.13` to `1.24`, matching this account's other Go repos, per
  `~/workspace/woodie/docs/COWORK.md`'s "Go version policy: target
  current Go, don't preserve legacy compatibility."

Made by inspection only, no Go toolchain in this sandbox -- needs a real
`make check` on the user's Mac to confirm clean.

## Done: Makefile added (build/lint/test/check)

`spec` had no Makefile at all -- matched `expect`'s shape (a pure library,
no `main` package, so `build` is a compile-only `go build ./...` sanity
check, no `install` target) over `humane`'s (which skips `build`
entirely) since it's slightly more complete and `expect` is `spec`'s
closest sibling in this pairing. `test` pipes through `gorderly -fd`:
`spec_test.go` already dogfoods `spec.Run` against `spec` itself, so this
renders as a real nested tree rather than a flat list. `check` stays
terse (silent on pass, full log on failure) via plain `go test`, so it
doesn't depend on `gorderly` being installed. README's new "Development"
section documents all four targets plus the no-`gorderly` fallback,
matching every other repo's precedent. No `.golangci.yml` added --
`expect`/`gorderly` don't have one either, just `golangci-lint run` with
defaults.

## Done: README refreshed around this fork's own positioning

Rewrote the opening to point explicitly at
[upstream's own README](https://github.com/sclevine/spec/blob/master/README.md)
and state directly why this fork exists (naming consistency, upstream
dormancy). Dropped the Travis/GoDoc badges -- both pointed at upstream's own
`sclevine/spec` CI/pkg.go.dev entries, not this fork's, so they were
misleading rather than just stale.

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
naming. The example now uses `it.BeforeEach`/`it.AfterEach`, updated in
the same pass that shipped the rename. `spec_test.go`/`options_test.go` (upstream's own test suite) still use
`when` -- left untouched for now, since rewriting real test code is a
bigger, riskier change than a README example; worth a separate pass if
wanted.

## Done: flipped which word is the base name -- `context` now, not `describe`

Reversed the direction recorded above (`describe` at the top level,
`context := describe` aliased). Prompted by a real discussion: some
suites in this account always use `context`, some mix `context` and
`describe`, matching how the wider RSpec community never settled on one
rule either (`context` has always been a plain alias for `describe`,
never a stricter or looser one). One argument for a direction, not a
requirement: Ginkgo's own source defines `Context` as `var Context =
Describe` -- `Describe` is the base symbol, `Context` its alias. Since
this fork deliberately ports Ginkgo's `JustBeforeEach` naming, matching
Ginkgo's own alias direction rather than inventing a reversed one seemed
worth doing too. Flipped: the raw `spec.Run`/`spec.G` parameter is now
named `context`, and `describe := context` is declared only in a suite
that actually calls `describe(...)` -- e.g. naming the method under test,
RSpec-style (`describe("#divide", ...)`), with `context` nested inside
for the conditions that vary.

**Caught before shipping:** the first pass at this flip declared
`describe := context` in the README's own Quick example and
`JustBeforeEach` example without ever calling `describe(...)` anywhere in
either body -- an alias declared and never used, which is not just
unhelpful documentation, it's a genuine Go compile error (unused local).
Fixed by giving each example a real reason to use both words: the Quick
example's "some slow things happen" group became `describe(...)` instead
of `context(...)` (a distinct capability, not a condition on the same
object); the `JustBeforeEach` example's `#divide` group became
`describe(...)` (naming the method under test, exactly the RSpec idiom
`describe` exists for), with `context("dividing evenly", ...)`/
`context("dividing with a remainder", ...)` nested inside for the actual
conditions. General rule going forward, everywhere this convention gets
documented or written: never declare an alias a suite doesn't call.

Rolled out account-wide in the same pass as the `BeforeEach`/`AfterEach`/
`JustBeforeEach` consumer bump: `gorderly` (every real file only calls
`context(...)`, so the alias is dropped entirely, not flipped), `expect`
and `humane`'s `time_test.go` (both call `describe(...)` for real, alias
kept), `lambada` (every file across both packages calls `describe(...)`
for real). See each repo's own `docs/COWORK.md` for the file-by-file
detail.

## Done: module path renamed to `github.com/woodie/spec` (`v0.3.0`)

The open question flagged above (no Go Reference badge while `go.mod`
still declared `module github.com/sclevine/spec`) is settled. Filed as
[woodie/spec#3](https://github.com/woodie/spec/issues/3), shipped through
PR #4 and tagged `v0.3.0`. Why:
`sclevine` never responded to the Jul 18 email, and public writing about
this pairing is now in scope, which makes the current state (clone the
fork, open `go.mod`, see upstream's own module path) a real point of
confusion for a first-time reader in a way it wasn't for four consumers
all owned by the same person.

The change fixed the internal self-imports (`report/log.go`,
`report/terminal.go`, `spec_test.go`, `options_test.go`) and continued the
existing version sequence (`v0.3.0`, not a symbolic `v1.0.0`). All four
consumers (`expect`, `gorderly`, `humane`, `lambada`) now plainly
`require github.com/woodie/spec v0.3.0`, with no `replace` left. For
`gorderly` this was a real unlock, not just cleanup: the `go install`/
`replace` conflict that forced its `v0.4.1` revert to plain upstream went
away, so it could adopt `BeforeEach`/`AfterEach` in its own suite too.

Explicitly not a build-safety fix -- worth being honest about this since
it's the instinct that started the thread. Every current consumer already
resolves `spec` through the `replace` directive straight to
`github.com/woodie/spec`'s own source and has `go.sum` checksums under
that path already; `sclevine` deleting the upstream repo today wouldn't
break anything already pinned. What the rename actually buys is
first-time adoption (`go get github.com/woodie/spec` just working,
pkg.go.dev indexing the real path).
