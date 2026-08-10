**Rename `spec`'s module declaration from `github.com/sclevine/spec` to
`github.com/woodie/spec`.** Filed as
[woodie/spec#3](https://github.com/woodie/spec/issues/3); this is the
fuller implementation plan the issue itself doesn't carry, in the same
spirit as `ISSUE_DRAFT_justbeforeeach.md` sketching the hook rename before
that one shipped.

## Why now, not just "someday"

Flagged as an open question back when CI/badges were added (see this
repo's own `docs/COWORK.md`, "Deliberately did **not** add a Go
Reference/pkg.go.dev badge" section) and left undecided because it was
"bigger than this session's scope." Two things changed since:

- Emailed `sclevine` directly on Jul 18 (cc'ing other Ginkgo/Gomega-
  adjacent contributors); no response after 13 days, and upstream's own
  issue tracker has issue creation restricted. Reasonable to stop waiting.
- Public writing is now in scope. A blog post pointing at this fork is a
  different audience than "every consumer is one person's own repos" --
  someone cloning `woodie/spec` fresh and opening `go.mod` today sees
  `module github.com/sclevine/spec`, which reads like a mistake or a
  half-finished fork, not a deliberate one.

**Not the driver: build safety.** Worth being precise about this, since
it's the instinct that started this whole thread. If `sclevine` deleted
`sclevine/spec` outright today, nothing currently pinned would break --
`expect`, `humane`, and `lambada` all resolve `spec` via `replace
github.com/sclevine/spec => github.com/woodie/spec v0.2.0`, and a
`replace` target is fetched straight from its own source; Go never goes
back to query the replaced-away original. `go.sum` in all three already
holds checksums under `github.com/woodie/spec v0.2.0`, not
`sclevine/spec`. Existing pins are already safe. What renaming actually
buys is adoption and honesty for anyone starting fresh -- `go get
github.com/woodie/spec` just working, pkg.go.dev indexing the real code,
and no `replace`-directive knowledge required to use it -- plus one real
technical unlock for `gorderly` (below).

## What actually has to change

### `spec` itself

- `go.mod`: `module github.com/sclevine/spec` -> `module
  github.com/woodie/spec`.
- Internal self-imports, since `spec` imports its own subpackage by full
  path: `report/log.go:7` and `report/terminal.go:8` (`import
  "github.com/sclevine/spec"`) both need to become
  `"github.com/woodie/spec"`. Same for the two files that dogfood `spec`
  against itself, `spec_test.go:8-9` (`"github.com/sclevine/spec"` and
  `"github.com/sclevine/spec/report"`) and `options_test.go:10`.
- `README.md`: intro line ("This is `woodie`'s fork of `sclevine/spec`")
  stays -- it's attribution, not the import path -- but this is the point
  where the deliberately-omitted Go Reference badge (see `docs/COWORK.md`)
  can finally be added for real, since pkg.go.dev will index the actual
  path being used.
- `LICENSE` stays untouched. Apache-2.0's copyright notice (Stephen
  Levine, 2017) is retained regardless of module path -- renaming the
  import path isn't a licensing event.
- Tag the rename as its own version. Recommend continuing the existing
  sequence (`v0.3.0`) rather than jumping to `v1.0.0` -- nothing about the
  API changed, only where the module lives, and this account's convention
  elsewhere (`expect`, `humane`) is small deliberate version bumps, not
  symbolic major-version ceremony for a plumbing change.
- Not a flag day: `v0.1.0`/`v0.2.0` are already tagged and immutable --
  their `go.mod` still says `module github.com/sclevine/spec` forever,
  and any consumer that never bumps past `v0.2.0` keeps working exactly as
  it does today, `replace` and all. The rename only takes effect for
  whoever explicitly moves to the new version.
- Separate, optional, not blocking: GitHub's own "forked from
  `sclevine/spec`" banner (visible on the repo page and on issue #3) is a
  platform-level relationship, independent of what `go.mod` declares --
  renaming the module doesn't remove it. Detaching the fork relationship
  entirely is a GitHub support request, worth doing only if the banner
  itself becomes something you want gone before publishing, not part of
  this plan.

### Consumers -- `expect`, `humane`, `lambada` (library-shaped, already on the fork)

All three currently carry the same two lines:

```
require github.com/sclevine/spec v1.4.0
replace github.com/sclevine/spec => github.com/woodie/spec v0.2.0
```

Once `spec` v0.3.0 is tagged under its new path, each becomes a single
plain line, no `replace` at all:

```
require github.com/woodie/spec v0.3.0
```

Per repo:

- Update every test file's import from `"github.com/sclevine/spec"` to
  `"github.com/woodie/spec"` -- `expect/expect_test.go`;
  `humane/human_size_test.go`, `humane/distance_in_time_test.go`,
  `humane/time_ago_test.go`; `lambada/cmd/lambada-mta/attachments_test.go`,
  `lambada/cmd/lambada-web/{middleware,scanfiles,server,main}_test.go`.
- `go mod tidy` to regenerate `go.sum` (drops the `sclevine/spec v1.4.0`
  require-only entry, keeps/refreshes the `woodie/spec` checksum that's
  already there from the `replace`).
- `expect/README.md`'s Setup section: the paragraph added for exactly
  this transition ("The import path stays `github.com/sclevine/spec` --
  `woodie/spec` keeps upstream's own module path unchanged...") becomes
  wrong the moment the path changes -- rewrite it to say the import *is*
  `github.com/woodie/spec` now, drop the `replace` snippet.
- `gorderly/docs/FRAMEWORK.md`'s closing note (the same explanation, plus
  the full-suite example's `import "github.com/sclevine/spec"` at
  line 381) needs the identical treatment.
- `make check`/`go test` on each, same "publish first, then bump each
  consumer's pin, confirm clean one at a time" sequence this account
  already uses for shared-library bumps (see any of the three's own
  `docs/COWORK.md`, "Shared libraries across sibling repos").

### `gorderly` -- different starting point, real unlock available

`gorderly` isn't on the fork today at all -- its `go.mod` has plain
`github.com/sclevine/spec v1.4.0`, no `replace`, and its test suite still
calls `it.Before`/`it.After`. Not an oversight: `v0.4.0` shipped with the
`replace` directive and broke `go install github.com/woodie/gorderly@latest`
outright, because Go rejects any `replace` in a `go.mod` being treated as
the main module under `go install pkg@version`. `v0.4.1` reverted to plain
upstream specifically to fix that (documented in `gorderly`'s own
`docs/COWORK.md` and `docs/FRAMEWORK.md`'s "Exception" section).

Renaming `spec`'s module removes the reason for that exception. The
`go install` restriction is specifically about `replace` directives --
once `github.com/woodie/spec` is a real, independent module, `gorderly`
can add a plain `require github.com/woodie/spec v0.3.0` with **no
`replace`**, which `go install` has no problem with. That means:

- `gorderly` can finally adopt `it.BeforeEach`/`it.AfterEach` in its own
  test suite (`parse_test.go`, `version_test.go`, `render_test.go`,
  `main_test.go`), closing the gap `FRAMEWORK.md` currently documents as a
  known exception.
- `FRAMEWORK.md`'s "Exception" section either gets deleted or rewritten
  past-tense once this ships -- confirm `go install
  github.com/woodie/gorderly@latest` actually works post-change before
  declaring it closed, not just by inspection.

This is the one piece of this plan that's a genuine improvement beyond
"clean up the naming," worth calling out on its own if it ends up in a
blog post.

## Suggested order

1. `spec`: rename, fix internal imports, tag `v0.3.0`. Confirm `make
   check` clean on the real toolchain before tagging (per this repo's own
   `docs/COWORK.md`, nothing here has been run against a real Go
   toolchain from this sandbox).
2. `expect` first among consumers -- smallest surface (one test file), and
   it's the one whose README already carries the explanatory paragraph
   that needs rewriting, so it's a good template for the other two.
3. `humane`, then `lambada` -- same mechanical change, more files.
4. `gorderly` last, since it's the odd one out (adding the dependency for
   the first time rather than dropping a `replace`) and the one place a
   real `go install` smoke test is worth running before calling it done.
5. Once all four are clean: revisit whether `woodie/docs/COWORK.md`'s
   "Shared libraries across sibling repos" section wants a line about
   this pattern specifically (a `replace`-based fork later graduating to
   its own module path) -- it's documented per-repo above but not yet as
   a generalized lesson.

## Open, not decided here

- Exact next version number (`v0.3.0` recommended above, not settled).
- Whether to also request GitHub detach the fork relationship, and if so,
  before or after publishing anything.
