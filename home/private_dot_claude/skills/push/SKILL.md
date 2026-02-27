---
name: push
description: Push the current branch to all configured remotes. Use when the user explicitly asks to push changes.
disable-model-invocation: true
allowed-tools: Bash, Read
---

Push the current branch to all configured remotes. Work through each step in order and stop immediately if a hard stop condition is met.

## Step 1 — Understand current state

Run in parallel:
- `git branch --show-current` — current branch name
- `git status --short` — check for uncommitted changes
- `git remote -v` — list all configured remotes (dedup: each remote appears twice in -v output, use unique names only)

Then run: `git log @{u}..HEAD --oneline 2>/dev/null` — show unpushed commits. If this fails (no upstream set), note that this is the first push for this branch.

If there are **no configured remotes**, stop here and say so.

If **already up to date** (zero unpushed commits AND upstream exists), report "already up to date" and stop — do not push unnecessarily.

If the **working tree is dirty** (uncommitted changes exist), warn the user but do not block.

## Step 2 — Check push constraints from CLAUDE.md

Read the project CLAUDE.md for any remote or branch push constraints. Look for:
- Protected remotes on specific branches (e.g. "main is protected on GitLab")
- Required push order across remotes
- Branches that should never be pushed directly

Note any constraints — apply them in Step 3.

## Step 3 — Push to each remote

Push the current branch to each configured remote, in order (`origin` first, then others).

For each remote:

```bash
# First push to this remote for this branch (no upstream set):
git push -u <remote> <branch>

# Subsequent pushes (upstream already set):
git push <remote> <branch>
```

Rules:
- **Never use `--force` or `-f`** unless the user explicitly requested it in their message
- If a constraint from Step 2 blocks a specific remote on the current branch, **skip it** and explain why (do not error)
- If a push fails on one remote, report the error and **continue to the next remote** — do not stop entirely

## Step 4 — Report results

Show a clear summary:

```
Push Results:
- Branch: <branch>
- <remote> (<url>): ✓ pushed  /  ✗ failed: <error>  /  ⊘ skipped: <reason>
- Commits pushed: <n>
```

## Hard stops

- **No remotes configured** — stop immediately.
- **Force push** — never add `--force` or `-f` without explicit user instruction.
- **CLAUDE.md hard block** — if CLAUDE.md explicitly forbids pushing the current branch anywhere, stop and explain.
