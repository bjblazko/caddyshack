---
description: File a new CaddyShack release: update CHANGELOG, tag, and push to trigger GitHub Actions binary builds
---

This command guides you through creating a new release. When the tag is pushed, GitHub Actions automatically builds binaries for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, and windows/amd64 and attaches them to the GitHub release.

## Steps

### 1. Get the version number

Ask the user: "What version number for this release? (e.g., 0.2.0 — no leading 'v')"

Validate that the input looks like a semver string (digits and dots only, e.g., `1.2.3`). The git tag will be `v{version}`.

### 2. Check for a clean working tree

Run:
```
git status --porcelain
```

If there is any output (staged, unstaged, or untracked files that should not be there), **STOP** and tell the user:

> There are uncommitted changes. Please commit or stash all changes before releasing, then run /release again.

### 3. Inspect the [Unreleased] section

Read CHANGELOG.md. Find everything between `## [Unreleased]` and the next `## [` heading. Count how many `### Category` lines are present inside that block.

- **Zero entries:** Ask the user: "The [Unreleased] section is empty. Do you want to release anyway (with no changelog entry), or add notes first with /update-docs?"
  - If they want to proceed without notes, continue.
  - If they want to add notes first, stop here.

- **Duplicate category headers** (e.g. two `### Added` blocks): **STOP** and tell the user:

  > The [Unreleased] section has duplicate category headers: [list the duplicates]. Please merge the duplicate sections so each category appears at most once, then run /release again.

- **One or more category headers with no duplicates (or zero entries and user confirmed):** Proceed.

### 4. Find the previous version

Look at the existing version entries in CHANGELOG.md (lines matching `## [x.y.z]`). The most recent one is the previous version. You will need it for the comparison URL.

### 5. Update CHANGELOG.md

Make the following changes in one edit:

a. Replace the `## [Unreleased]` line with:
```
## [Unreleased]

## [{version}] - {today's date in YYYY-MM-DD format}
```

b. At the bottom of the file, update the `[Unreleased]` reference link to point to `v{version}...HEAD`:
```
[Unreleased]: https://github.com/bjblazko/caddyshack/compare/v{version}...HEAD
```

c. Add a new reference link for the new version:
```
[{version}]: https://github.com/bjblazko/caddyshack/compare/v{previous}...v{version}
```
If there is no previous version, use:
```
[{version}]: https://github.com/bjblazko/caddyshack/releases/tag/v{version}
```

### 6. Stage and commit

```
git add CHANGELOG.md
git commit -m "chore: release v{version}"
```

### 7. Create an annotated tag

```
git tag -a v{version} -m "Release v{version}"
```

### 8. Push commit and tag

```
git push origin main
git push origin v{version}
```

### 9. Confirm success

Tell the user:

> Release v{version} is underway. Tag `v{version}` has been pushed. GitHub Actions will build and upload binaries for 5 platforms. Monitor progress at: https://github.com/bjblazko/caddyshack/actions
