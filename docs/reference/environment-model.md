# Environment Model

Templates use two variables: **platform** (auto-detected) and **context** (user-selected at `chezmoi init`).

## Platform Detection

Detected automatically — never prompted:

| Value | Detected when |
|---|---|
| `distrobox` | `$DISTROBOX_ENTER_PATH` is set |
| `wsl` | `/proc/version` contains "microsoft" |
| `aurora` | `/etc/os-release` contains "aurora" |
| `unknown` | none of the above |

## Context Values

Chosen by the user at `chezmoi init`:

| Context | Platform | Use case |
|---|---|---|
| `personal` | aurora, wsl | Personal laptop host (aurora) or personal WSL instance |
| `work-eam` | wsl or distrobox | EAM work projects — work-isolated WSL or dev container |
| `work-<name>` | distrobox | Any other work context |
| `personal-<project>` | distrobox | Project-scoped dev (Bun, Playwright, native op) |

## What Each Context Gets

Features vary by both context and platform. Bootstrap installs tools on
WSL/Distrobox; Aurora installs most tools via brew (see aurora.md step 1.9).
Rows marked with a platform note only apply on that platform.

| Feature | personal | personal-\<project\> | work-\<name\> |
|---|---|---|---|
| 1Password SSH agent | host socket (aurora/distrobox), npiperelay (wsl) | host socket | host socket (aurora/distrobox), npiperelay (wsl) |
| glab | ✓ (aurora via brew, distrobox via bootstrap) | ✓ (distrobox) | — |
| kubectl | ✓ (aurora via brew, wsl/distrobox via bootstrap) | — | ✓ (wsl/distrobox) |
| terraform | — | — | ✓ (wsl/distrobox) |
| AWS CLI | — | — | ✓ (wsl/distrobox) |
| ansible | ✓ (distrobox) | — | — |
| op CLI | ✓ (aurora via rpm-ostree) | ✓ (distrobox native) | — |
| NVM | ✓ (wsl/distrobox) | ✓ (distrobox) | ✓ (wsl/distrobox) |
| bun | ✓ (wsl) | ✓ (wsl/distrobox) | ✓ (wsl) |
| atuin | ✓ | ✓ | ✓ |
| OTel telemetry | ✓ | ✓ | ✓ |
| Claude Code | ✓ | ✓ | ✓ |
| setup-creds | ✓ (distrobox only) | ✓ (distrobox only) | ✓ (distrobox only) |

## Adding a New Context

**New work context:**
1. Add container to `containers/distrobox.ini`
2. Add job-specific aliases in `home/dot_zshrc.tmpl` under `hasPrefix .context "work-"`
3. Run: `uv run python scripts/distrobox_setup.py work-<name> --work-email you@company.com`

**New personal project:**
1. Add container to `containers/distrobox.ini`
2. Run: `uv run python scripts/distrobox_setup.py personal-<name> --personal-email you@email.com`

---

For isolated AI sandboxes (vibe-coded projects), see `ai-sandbox` in `bin/ai-sandbox`.
