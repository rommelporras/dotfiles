# Claude Code Plugin Dependencies Design

Date: 2026-03-05

## Problem

Claude Code plugins (Context7, Playwright, Superpowers, episodic-memory) require
runtime dependencies (Node.js, 1Password access) that aren't consistently available
across all environments (Aurora, WSL, Distrobox, AI sandbox). Plugin and MCP setup
is manual and per-machine.

Additionally, `~/.claude/` config is managed by chezmoi with templating, but the
files (settings.json, hooks, skills, agents) are identical across environments.
A separate `claude-config` git repo already exists for this purpose and is simpler.

## Decisions

1. **claude-config repo owns `~/.claude/`** — git clone + symlinks, not chezmoi
2. **Universal CLAUDE.md** — no per-platform templating; one file covers all platforms
3. **NVM for Node.js** on WSL + Distrobox (Aurora uses brew)
4. **Native installer** for Claude Code on WSL + Distrobox (Aurora has it via ujust)
5. **setup-creds handles plugin/MCP setup** — co-located with credential setup
6. **episodic-memory enabled globally** in settings.json
7. **AI sandbox Containerfile** switches to native Claude Code installer

## Architecture

### ~/.claude/ Management

Source: `~/personal/claude-config` (git repo, GitHub remote)

```
~/personal/claude-config/
├── CLAUDE.md          ──symlink──▶  ~/.claude/CLAUDE.md
├── settings.json      ──symlink──▶  ~/.claude/settings.json
├── hooks/             ──symlink──▶  ~/.claude/hooks/
├── skills/            ──symlink──▶  ~/.claude/skills/
├── agents/            ──symlink──▶  ~/.claude/agents/
└── .gitignore         (excludes runtime state)
```

Runtime state (plugins/, projects/, history.jsonl, credentials.json) stays in
`~/.claude/` as real files — untouched by symlinks.

**Dotfiles repo change:** Delete `home/private_dot_claude/` entirely. Remove the
old symlink cleanup block from bootstrap (lines 24-31).

### Distrobox Symlink Path

Inside distrobox, `$HOME` is `~/.distrobox/<context>/`. The claude-config repo
lives on the host at `/home/<user>/personal/claude-config`. Symlinks use the
absolute host path (containers see the host filesystem via distrobox mounts).

### CLAUDE.md (Universal)

Replaces the old templated `CLAUDE.md.tmpl` with a single file covering all platforms:

```markdown
## Environment

Environment is auto-detected. Key constraints per platform:

- **Aurora DX** (immutable Fedora): No `apt-get`, no `chsh`. Use `rpm-ostree` or `brew`.
  Shell is zsh via Ptyxis custom command. 1Password SSH Agent at `~/.1password/agent.sock`.
- **WSL2**: No `op` access in this terminal. Windows Chrome for browser automation.
- **Distrobox**: No `op` access — use `distrobox-host-exec op` if needed.
  `$HOME` is container-local, not the host home.

All platforms:
- Self-hosted GitLab is the primary remote; use `glab` CLI for GitLab operations.
- Work projects may use GitHub instead.
- **Never run `op` commands** — generate them and ask the user to run manually.
```

### Claude Code Installation

| Platform   | Method                                           | In bootstrap? |
|------------|--------------------------------------------------|---------------|
| Aurora     | Pre-installed via `ujust bbrew` (brew cask)      | No — skip     |
| WSL        | `curl -fsSL https://claude.ai/install.sh \| bash` | Yes           |
| Distrobox  | `curl -fsSL https://claude.ai/install.sh \| bash` | Yes           |
| AI sandbox | `curl -fsSL https://claude.ai/install.sh \| bash` | Containerfile |

Guard: `command -v claude` — skip if already present.

### Node.js for Plugin Runtime

| Platform   | Method                          | Current state       |
|------------|---------------------------------|---------------------|
| Aurora     | `brew install node@24`          | Already in bootstrap|
| WSL        | NVM v0.40.1                     | Already in bootstrap|
| Distrobox  | NVM v0.40.1                     | **NEW — add**       |

Change the NVM install condition from `{{ if eq .platform "wsl" }}` to
`{{ if or (eq .platform "wsl") (eq .platform "distrobox") }}`.

### Plugin & MCP Setup (setup-creds)

Add to `setup-creds` script after the existing credential blocks:

```bash
# ─── Claude Code plugins & MCP ───────────────────────────
if command -v claude &>/dev/null; then
    echo "Setting up Claude Code plugins..."

    # Marketplaces
    claude plugin marketplace add anthropics/claude-plugins-official 2>/dev/null || true
    claude plugin marketplace add obra/superpowers-marketplace 2>/dev/null || true

    # Plugins (user scope)
    claude plugin install context7@claude-plugins-official --scope user 2>/dev/null || true
    claude plugin install superpowers@superpowers-marketplace --scope user 2>/dev/null || true
    claude plugin install episodic-memory@superpowers-marketplace --scope user 2>/dev/null || true

    # Context7 MCP (needs 1Password API key)
    OP_CMD="op"
    if [ -n "${DISTROBOX_ENTER_PATH:-}" ]; then
        OP_CMD="distrobox-host-exec op"
    fi
    CONTEXT7_KEY=$($OP_CMD read 'op://Kubernetes/Context7/api-key' --no-newline 2>/dev/null)
    if [ -n "$CONTEXT7_KEY" ]; then
        claude mcp add --scope user --transport http context7 \
            https://mcp.context7.com/mcp \
            --header "CONTEXT7_API_KEY: $CONTEXT7_KEY"
        echo "  Context7 MCP registered"
    else
        echo "  WARNING: Could not read Context7 API key from 1Password"
    fi
fi
```

### settings.json Update

Add episodic-memory to enabledPlugins in claude-config:

```json
"enabledPlugins": {
    "superpowers@superpowers-marketplace": true,
    "context7@claude-plugins-official": true,
    "playwright@claude-plugins-official": true,
    "episodic-memory@superpowers-marketplace": true
}
```

### AI Sandbox Containerfile

Replace `npm install -g @anthropic-ai/claude-code` with:

```dockerfile
RUN curl -fsSL https://claude.ai/install.sh | bash
```

Keep Node.js install (still needed for other tools in the sandbox).

### distrobox-setup.sh Update

After chezmoi apply, before setup-creds, add claude-config symlink setup:

```bash
# Link claude-config into container's ~/.claude/
CLAUDE_CONFIG="${REPO_DIR%/*}/claude-config"
distrobox enter "$container" -- sh -c "
    CLAUDE_CONFIG='$CLAUDE_CONFIG'
    if [ -d \"\$CLAUDE_CONFIG\" ]; then
        mkdir -p \"\$HOME/.claude\"
        for item in CLAUDE.md settings.json hooks skills agents; do
            ln -sf \"\$CLAUDE_CONFIG/\$item\" \"\$HOME/.claude/\$item\"
        done
        echo 'Linked ~/.claude/ → claude-config'
    else
        echo 'WARNING: claude-config not found at $CLAUDE_CONFIG'
    fi
"
```

## Testing Plan

### Environments to test (in order)

| # | Environment        | Setup command                    |
|---|--------------------|----------------------------------|
| 1 | Aurora host        | `chezmoi apply -v`               |
| 2 | Distrobox personal | `distrobox-setup.sh personal`    |
| 3 | Distrobox work-eam | `distrobox-setup.sh work-eam`    |
| 4 | Distrobox sandbox  | `distrobox-setup.sh sandbox`     |
| 5 | AI sandbox         | `podman build -f containers/Containerfile.ai-sandbox` |

### Verification checklist (per environment)

```bash
# 1. Claude Code binary
claude --version

# 2. Node.js available
node --version

# 3. Symlinks correct
ls -la ~/.claude/CLAUDE.md ~/.claude/settings.json ~/.claude/hooks ~/.claude/skills ~/.claude/agents

# 4. Symlinks resolve (not dangling)
cat ~/.claude/CLAUDE.md | head -3

# 5. Plugin marketplaces registered
claude plugin marketplace list

# 6. Plugins installed
claude plugin list

# 7. Context7 MCP registered
claude mcp list

# 8. Episodic memory functional (non-sandbox only)
# Start Claude Code, verify episodic-memory tools are available
```

### Sandbox-specific checks

```bash
# Verify NO claude config, NO plugins, NO MCP
ls ~/.claude/ 2>&1        # should not exist or be empty
claude plugin list 2>&1   # should show nothing
claude mcp list 2>&1      # should show nothing
```

### Rollback

If something breaks:
1. Remove symlinks: `rm ~/.claude/{CLAUDE.md,settings.json,hooks,skills,agents}`
2. Re-run `chezmoi apply` (old chezmoi-managed files restore)
3. Revert bootstrap changes in dotfiles repo

## Files Changed

### Dotfiles repo (this repo)
- `home/private_dot_claude/` — **DELETE** (entire directory)
- `home/run_once_before_bootstrap.sh.tmpl` — add claude-config clone/symlink, Claude Code install, expand NVM to distrobox, update post-install instructions
- `home/dot_local/bin/executable_setup-creds.tmpl` — add plugin/MCP setup block
- `scripts/distrobox-setup.sh` — add claude-config symlink step
- `containers/Containerfile.ai-sandbox` — switch to native installer
- `home/.chezmoiignore` — remove `.claude/` entries that were for chezmoi management
- `CLAUDE.md` (project) — update to reflect new architecture

### claude-config repo (separate)
- `CLAUDE.md` — replace "Personal Environment (WSL2)" with universal Environment section
- `settings.json` — add episodic-memory to enabledPlugins
