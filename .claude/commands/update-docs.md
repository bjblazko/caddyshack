---
description: Update CHANGELOG.md [Unreleased] section and README.md after making changes
---

This command helps you document recent changes in CHANGELOG.md and README.md.

## Steps

1. Ask the user: "What changed since the last release? Please describe new features, fixes, or other notable changes."

2. Based on the response, determine the correct Keep-a-Changelog category:
   - **Added** — new features
   - **Changed** — changes to existing functionality
   - **Deprecated** — features that will be removed in a future release
   - **Removed** — features that were removed
   - **Fixed** — bug fixes
   - **Security** — security fixes

3. Open CHANGELOG.md and add entries under the `## [Unreleased]` section.
   - Group entries under the appropriate `### Category` header (create it if it does not exist)
   - Write bullets in present-tense imperative style: "Add X", "Fix Y", "Remove Z"
   - **Note:** accumulating multiple categories in [Unreleased] is fine. The single-category-per-release rule is enforced at release time by the `/release` command, not here.

4. If any new features were added or existing ones changed, ask: "Should the README be updated to reflect these changes?"
   - If yes, update the relevant sections of README.md (Features list, Architecture section, etc.)

5. Show the user a diff summary of all changes made to CHANGELOG.md and README.md.
