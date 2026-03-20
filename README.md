# dotfiles

chezmoi-managed dotfiles for consistent dev environments across WSL2, Aurora DX (immutable Fedora), and Distrobox containers. Includes AI sandboxing for vibe-coded projects and `dotctl` for monitoring dotfiles status across machines.

## Quick Start

| Platform | Guide | What it does |
|---|---|---|
| **Aurora DX** | [docs/setup/aurora.md](docs/setup/aurora.md) | Host setup: devmode, brew, zsh, 1Password, chezmoi |
| **Distrobox** | [docs/setup/distrobox.md](docs/setup/distrobox.md) | Container setup: create, bootstrap, credential seeding |
| **WSL2** | [docs/setup/wsl2.md](docs/setup/wsl2.md) | Ubuntu 24.04 instance with npiperelay SSH bridge |

All platforms run `chezmoi init --apply` — it detects the platform automatically and prompts for context.

## Environment Model

Two variables drive all template logic:

- **`platform`** (auto-detected): `aurora`, `wsl`, `distrobox`
- **`context`** (user-selected): `personal`, `personal-<project>`, `work-<name>`

| Platform | Context | SSH Agent | Key tools |
|---|---|---|---|
| `aurora` | `personal` | 1Password native | Host OS, brew, immutable |
| `wsl` | `personal` | 1Password via npiperelay | NVM, Bun, personal creds |
| `wsl` | `work-eam` | 1Password via npiperelay | NVM, Bun, work + personal creds |
| `distrobox` | `personal` | 1Password host socket | glab, ansible, kubectl, Atuin |
| `distrobox` | `personal-<project>` | 1Password host socket | glab, Bun, Playwright, native op |
| `distrobox` | `work-<name>` | 1Password host socket | terraform, AWS CLI, kubectl, Atuin |

See [docs/reference/environment-model.md](docs/reference/environment-model.md) for the full matrix.

## AI Sandbox

Run AI agents in isolated, project-scoped Podman containers. Each project gets a
persistent home at `~/.sandbox/<project>/` with no host filesystem access.

```bash
ai-sandbox fintrack --shell                    # interactive shell
ai-sandbox fintrack claude -- --dangerously-skip-permissions  # run claude
ai-sandbox fintrack --services up              # start compose services
ai-sandbox fintrack --destroy                  # remove everything
ai-sandbox --list                              # list all projects
```

**What you get on first shell start** (all automated):
- 1Password SSH agent for `git clone`/`push`
- IDE forwarding: `code .`, `antigravity`, `agy` open on the host
- Browser forwarding for OAuth flows (`BROWSER` env var)
- Atuin login (via podman secrets)
- Claude Code with all plugins + Context7 MCP
- Claude config (CLAUDE.md, rules, hooks, skills, agents) linked from host

**One-time secret setup** (run on host):
```bash
op read "op://Kubernetes/Atuin/personal-password" | tr -d '\n' | podman secret create atuin_password -
op read "op://Kubernetes/Atuin/encryption-key" | tr -d '\n' | podman secret create atuin_key -
op read "op://Dev/Anthropic/credential" | tr -d '\n' | podman secret create anthropic_key -
op read "op://Dev/Context7/api-key" | tr -d '\n' | podman secret create context7_key -
```

See [docs/reference/credentials.md](docs/reference/credentials.md) for all credential paths.

## Day-to-Day

```bash
chezmoi diff                              # preview what would change
chezmoi apply -v                          # apply changes
chezmoi edit ~/.zshrc && chezmoi apply    # edit a managed file
```

Inside distrobox containers, chezmoi source is symlinked to the host repo — uncommitted changes apply immediately with `chezmoi apply`.

## dotctl

Monitor dotfiles status across all machines. Pushes metrics to a homelab OTel Collector.

```bash
make install            # build + install to ~/.local/bin/
make install-systemd    # enable auto-collect every 10 minutes

dotctl status --live    # local dashboard, no cluster needed
dotctl status           # query Prometheus + Loki
dotctl collect          # push metrics to OTel Collector
```

See [docs/reference/dotctl.md](docs/reference/dotctl.md) for full reference.

## Repository Layout

```
dotfiles/
├── home/              chezmoi source dir → maps to ~/
│   ├── .chezmoi.toml.tmpl       Interactive prompts (chezmoi init)
│   ├── .chezmoiignore           Per-environment file skipping
│   ├── dot_zshrc.tmpl           Shell config (templated)
│   ├── run_once_before_bootstrap.sh.tmpl  First-run setup
│   └── dot_local/bin/           User scripts (setup-creds)
├── containers/
│   ├── distrobox.ini            Container definitions (2 containers)
│   └── Containerfile.sandbox-base  AI sandbox image
├── scripts/           Distrobox + sandbox automation (Python, via uv)
├── bin/               CLI tools (ai-sandbox)
├── dotctl/            Go CLI (collect, status, status --live)
└── docs/
    ├── setup/         Per-platform setup guides
    ├── reference/     CLI ref, environment model, credentials
    └── architecture/  Design decisions
```

## Testing

```bash
# dotctl unit tests (29 tests)
make test

# Distrobox integration tests — destroys + recreates containers (59 assertions)
uv run python scripts/test_distrobox_integration.py --all
uv run python scripts/test_distrobox_integration.py personal

# Sandbox integration tests (51 assertions)
uv run python scripts/test_sandbox_integration.py
uv run python scripts/test_sandbox_integration.py --skip-build  # use existing image
```

## License

MIT. See [LICENSE](LICENSE).
