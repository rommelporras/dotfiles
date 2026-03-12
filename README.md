# dotfiles

chezmoi-managed dotfiles for consistent dev environments across WSL2, Aurora DX, and Distrobox containers. Includes `dotctl` — a CLI for monitoring dotfiles status across all machines.

## Quick Start

Choose your platform:

- **WSL2** — [docs/setup/wsl2.md](docs/setup/wsl2.md)
- **Aurora DX** — [docs/setup/aurora.md](docs/setup/aurora.md)
- **Distrobox containers** — [docs/setup/distrobox.md](docs/setup/distrobox.md)

All platforms run `chezmoi init --apply` — it detects the platform automatically and asks for context.

## dotctl

Monitor dotfiles status across all machines.

```bash
# Install
cd ~/personal/dotfiles
make install          # ~/.local/bin/dotctl
make install-systemd  # auto-collect every 10 minutes

# Commands
dotctl status           # query Prometheus + Loki dashboard
dotctl status --live    # collect locally, no cluster needed
dotctl collect --verbose  # push metrics to OTel Collector
```

See [docs/reference/dotctl.md](docs/reference/dotctl.md) for full reference.

## Day-to-Day

```bash
chezmoi diff        # preview what would change
chezmoi apply       # apply changes
chezmoi update      # pull latest + apply
chezmoi edit ~/.zshrc && chezmoi apply  # edit a managed file
```

## AI Sandbox

Run AI agents in an isolated Podman container with no access to host credentials.

```bash
ai-sandbox claude -- --dangerously-skip-permissions  # code only
ai-sandbox --git claude -- --dangerously-skip-permissions  # code + git push
ai-sandbox --no-network gemini  # maximum containment
```

## Repository Layout

```
dotfiles/
├── dotctl/          — Go CLI (build, collect, status)
├── home/            — chezmoi source → maps to ~/
├── scripts/         — Distrobox setup + integration tests (Python)
├── containers/      — distrobox.ini + Containerfile.ai-sandbox
├── bin/             — ai-sandbox CLI
└── docs/
    ├── setup/       — per-platform setup guides
    ├── reference/   — CLI reference, environment model, credentials
    └── architecture/ — design decisions
```

## Environment Model

Two variables: **platform** (auto-detected) and **context** (chosen at `chezmoi init`).

| Platform | Context | Description |
|---|---|---|
| `aurora` | `personal` | Personal laptop host |
| `wsl` | `gaming` | Gaming desktop |
| `wsl` | `work-eam` | Work laptop |
| `distrobox` | `personal` | Personal dev container |
| `distrobox` | `personal-<project>` | Project-scoped (Bun, Playwright, native op) |
| `distrobox` | `work-<name>` | Work dev container |
| `distrobox` | `sandbox` | Clean experiment space, no credentials |

See [docs/reference/environment-model.md](docs/reference/environment-model.md) for full matrix.

## Testing

```bash
# dotctl unit tests
make test

# Distrobox integration tests (Aurora only)
uv run python scripts/test_distrobox_integration.py --all
uv run python scripts/test_distrobox_integration.py personal
```

## License

MIT. See [LICENSE](LICENSE).
