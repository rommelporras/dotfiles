Generic release workflow. Creates a semantic version tag, pushes, and creates a platform release (GitHub or GitLab). Projects with their own `.claude/commands/release.md` override this.

## Arguments

```
/release                      → Auto-determine version from commits
/release v1.2.0               → Explicit version, auto-generate title
/release v1.2.0 "Title Here"  → Explicit version AND title
```

## Step 1 — Check current state

Run in parallel:
- `git status` — must have a clean working tree
- `git branch --show-current` — identify current branch
- `git remote -v` — list configured remotes
- `git tag --sort=-v:refname | head -5` — recent tags

**Hard stop** if working tree is dirty — do not release with uncommitted changes.

Read the project CLAUDE.md for any release-specific rules (branch policy, required checks, special files to update). Follow those rules — they take precedence over this generic workflow.

## Step 2 — Validate branch

Determine the project's release branch:
- If CLAUDE.md specifies a release branch, use that
- Otherwise, use the current branch if it's `main` or `master`
- **Hard stop** if on any other branch without CLAUDE.md guidance

## Step 3 — Fetch tags and check for collision

```bash
git fetch --tags
```

If `$ARGUMENTS` specifies a version, verify that tag does not already exist locally or remotely. **Hard stop** on collision.

## Step 4 — Determine version

If version provided in `$ARGUMENTS`, use it directly.

Otherwise, auto-bump based on commits since the last tag:

```bash
git log $(git describe --tags --abbrev=0 2>/dev/null || git rev-list --max-parents=0 HEAD)..HEAD --oneline
```

| Commit pattern | Bump |
|---|---|
| `BREAKING CHANGE` in body or `!:` | Major (v1.0.0 → v2.0.0) |
| `feat:` | Minor (v1.0.0 → v1.1.0) |
| `fix:`, `docs:`, `chore:`, `refactor:`, `infra:` | Patch (v1.0.0 → v1.0.1) |

If no previous tags exist, default to `v0.1.0`.

## Step 5 — Generate title

If title provided in `$ARGUMENTS`, use it.

Otherwise, generate a short title (3-6 words) summarising the main changes.

Format: `v<VERSION> - <Short Title>` (regular hyphen, not em dash).

## Step 6 — Categorise commits for release notes

Get all commits since the last tag and group them:

| Category | Commit types |
|---|---|
| Features | `feat:` |
| Bug Fixes | `fix:` |
| Infrastructure | `infra:` |
| Documentation | `docs:` |
| Chores | `chore:`, `refactor:` |

Omit empty categories. Each entry is a single-line bullet from the commit subject.

## Step 7 — Check for CHANGELOG.md

If the project has a `CHANGELOG.md`:
1. Check if an entry for this version already exists — skip if so
2. Otherwise, prepend a new entry in Keep a Changelog format:

```markdown
## [v<VERSION>] - <YYYY-MM-DD>

### Added
- ...

### Fixed
- ...

### Changed
- ...
```

3. Commit the CHANGELOG update:
```bash
git add CHANGELOG.md
git commit -m "docs: update CHANGELOG for v<VERSION>"
```

If no CHANGELOG.md exists, skip this step entirely — do not create one.

## Step 8 — Show release plan and confirm

Present the full plan to the user:

```
Release Plan:
  Version:    v<VERSION>
  Title:      v<VERSION> - <Short Title>
  Branch:     <branch>
  Remote(s):  <list>
  Platform:   GitHub / GitLab
  Commits:    <n> since <last tag>
  CHANGELOG:  Updated / Skipped / Not present

Pre-release checks:
  ✓ Clean working tree
  ✓ On release branch
  ✓ No tag collision
  ✓ Remote tags fetched

Release notes preview:
  <categorised commit list>
```

**Wait for user confirmation — do NOT proceed without it.**

## Step 9 — Execute release

### Create annotated tag

```bash
git tag -a v<VERSION> -m "<tag annotation with version, title, and summary>"
```

### Push

Detect remotes and push to each:

```bash
git push <remote> <branch>
git push <remote> v<VERSION>
```

If multiple remotes exist, push to all. `origin` first, then others.

### Create platform release

Detect platform from remote URL:

**GitHub** (remote contains `github.com`):
```bash
gh release create v<VERSION> --title "v<VERSION> - <Title>" --notes "<release notes>"
```

**GitLab** (remote contains `gitlab`):
```bash
glab release create v<VERSION> --name "v<VERSION> - <Title>" --notes "<release notes>"
```

If both platforms have remotes, create releases on both.

## Step 10 — Report results

```
Release complete:
  Tag:       v<VERSION>
  Title:     v<VERSION> - <Title>
  Pushed:    <remote(s)>
  Release:   <URL(s)>
  CHANGELOG: Updated / Skipped
```

## Quality checklist

- [ ] On release branch (main/master or CLAUDE.md-specified)
- [ ] Working tree clean
- [ ] Remote tags fetched, no collision
- [ ] Version follows SemVer
- [ ] Release notes are categorical and specific
- [ ] Tag annotation has context
- [ ] User confirmed before execution
- [ ] CHANGELOG updated (if file exists)

## Hard stops

- **Dirty working tree** — never release with uncommitted changes
- **Tag collision** — never overwrite an existing tag
- **No user confirmation** — always wait for explicit approval
- **Force push** — never use `--force` for release pushes
- **CLAUDE.md overrides** — if project CLAUDE.md has release rules, follow those instead of this generic workflow
- **NO AI attribution** — no "Generated with Claude Code" or "Co-Authored-By" in notes, tags, or commits
