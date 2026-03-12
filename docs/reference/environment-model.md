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
| `personal` | aurora, wsl | Personal laptop host (aurora) or work laptop personal WSL (wsl) |
| `gaming` | wsl | Gaming desktop — personal projects |
| `work-eam` | wsl or distrobox | EAM work projects — work-isolated WSL or dev container |
| `work-<name>` | distrobox | Any other work context |
| `personal-<project>` | distrobox | Project-scoped dev (Bun, Playwright, native op) |
| `sandbox` | distrobox | Clean experiment space, no credentials |

## What Each Context Gets

| Feature | personal | personal-\<project\> | work-\<name\> | gaming | sandbox |
|---|---|---|---|---|---|
| 1Password SSH agent | host socket | host socket | host socket | Windows bridge | fallback |
| glab | ✓ | ✓ | — | — | — |
| kubectl | ✓ | — | ✓ | — | — |
| terraform | — | — | ✓ | — | — |
| AWS CLI | — | — | ✓ | — | — |
| ansible | ✓ | — | — | — | — |
| op CLI (native) | — | ✓ | — | — | — |
| bun | — | ✓ | — | — | — |
| atuin | ✓ | ✓ | ✓ | ✓ | — |
| OTel telemetry | ✓ | ✓ | ✓ | ✓ | — |
| Claude Code | ✓ | ✓ | ✓ | ✓ | — |
| setup-creds | ✓ | ✓ | ✓ | — | — |

## Adding a New Context

**New work context:**
1. Add container to `containers/distrobox.ini`
2. Add job-specific aliases in `home/dot_zshrc.tmpl` under `hasPrefix .context "work-"`
3. Run: `uv run python scripts/distrobox_setup.py work-<name> --work-email you@company.com`

**New personal project:**
1. Add container to `containers/distrobox.ini`
2. Run: `uv run python scripts/distrobox_setup.py personal-<name> --personal-email you@email.com`
