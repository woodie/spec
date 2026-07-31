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

## Planned: a `JustBeforeEach` hook

Ginkgo has [`JustBeforeEach`](https://onsi.github.io/ginkgo/#separating-creation-and-configuration-justbeforeeach);
`spec` doesn't. The account's own workaround is the `subject`-closure
pattern (documented in `gorderly`'s `docs/FRAMEWORK.md`, "Why `spec`, not
Ginkgo"/"The subject pattern") -- it gets to the same place but has to be
invoked explicitly at each call site rather than running automatically.
[kotest/kotest#952](https://github.com/kotest/kotest/issues/952) (filed
2019 against KotlinTest, closed with no resolution) shows the identical
gap recurring in a completely different BDD ecosystem, and cites Ginkgo's
`JustBeforeEach` as the model too -- not a `spec`-specific complaint.

Drafted a real issue (`ISSUE_DRAFT_justbeforeeach.md` in this repo's root,
not yet filed) proposing the feature, with a concrete implementation shape
sketched against `spec.go`'s actual hook mechanism (the `specHooks`/
`specHook` linked list, built via `hooks.next()`, run via
`specHooks.run`): a new `S.JustBefore` method, a `config.justBefore` flag,
a `justBefore []func()` slice on `specHook`, and `specHooks.run` collecting
every level's `justBefore` funcs while walking `before`s, then running all
of them in that same outermost-to-innermost order immediately before the
spec itself. Not yet filed, not yet implemented, not yet confirmed against
a real Go toolchain (no Go in this sandbox -- see `~/workspace/woodie/docs/COWORK.md`'s
"Working on unfamiliar stacks").

Next steps once the issue's filed: implement the hook on a real branch,
add spec coverage for it, confirm on the user's own Mac, then decide
whether `expect`/`gorderly`/`humane` adopt the fork (via `replace` or a
real `require` bump) for this feature specifically -- same "publish
first, then bump each consumer's pin" sequence as any other shared-library
change, per the cross-project doc's "Shared libraries across sibling
repos" section.
