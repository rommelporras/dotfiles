# Claude Code Plugin Dependencies Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make Claude Code plugins work reliably across Aurora, WSL, Distrobox, and AI sandbox by fixing dependency gaps and switching ~/.claude/ management from chezmoi to the claude-config git repo.

**Architecture:** claude-config repo (git clone + symlinks) owns all static ~/.claude/ files. Bootstrap installs Claude Code + Node.js. setup-creds handles plugin marketplace/install and Context7 MCP registration. Each environment gets identical config via symlinks to one source.

**Tech Stack:** chezmoi templates (Go text/template), bash scripts, claude CLI, NVM, 1Password CLI

---

### Task 1: Update claude-config repo (CLAUDE.md + settings.json)

**Files:**
- Modify: `~/personal/claude-config/CLAUDE.md` (replace lines 20-31)
- Modify: `~/personal/claude-config/settings.json` (line 65-67)

**Step 1: Update CLAUDE.md environment section**

Replace the "Personal Environment (WSL2)" section with the universal version:

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

**Step 2: Add episodic-memory to settings.json**

Add `"episodic-memory@superpowers-marketplace": true` to the `enabledPlugins` object.
Remove `"model": "sonnet"` line (claude-config has it, deployed settings.json doesn't — keep them in sync by removing it).

**Step 3: Verify files look correct**

Run: `cat ~/personal/claude-config/CLAUDE.md`
Run: `cat ~/personal/claude-config/settings.json | python3 -m json.tool`
Expected: valid JSON, 4 plugins in enabledPlugins, universal environment section in CLAUDE.md

**Step 4: Commit**

```bash
cd ~/personal/claude-config
git add CLAUDE.md settings.json
git commit -m "feat: universal CLAUDE.md and add episodic-memory plugin"
```

---

### Task 2: Delete private_dot_claude from dotfiles repo

**Files:**
- Delete: `home/private_dot_claude/` (entire directory — 10 files)

**Step 1: Verify no unique content will be lost**

Compare the files between claude-config and private_dot_claude to confirm they're equivalent:

Run: `diff ~/personal/claude-config/CLAUDE.md ~/personal/dotfiles/home/private_dot_claude/CLAUDE.md.tmpl`
Run: `diff ~/personal/claude-config/settings.json ~/personal/dotfiles/home/private_dot_claude/settings.json`
Run: `diff -rq ~/personal/claude-config/hooks/ ~/personal/dotfiles/home/private_dot_claude/hooks/`
Run: `diff -rq ~/personal/claude-config/skills/ ~/personal/dotfiles/home/private_dot_claude/skills/`
Run: `diff -rq ~/personal/claude-config/agents/ ~/personal/dotfiles/home/private_dot_claude/agents/`

Expected: differences are only the template conditionals in CLAUDE.md.tmpl and the `executable_` prefix on hook filenames (chezmoi convention). No unique content.

**Step 2: Delete the directory**

Run: `rm -rf ~/personal/dotfiles/home/private_dot_claude/`

**Step 3: Verify it's gone**

Run: `ls ~/personal/dotfiles/home/private_dot_claude/ 2>&1`
Expected: "No such file or directory"

**Step 4: Commit**

```bash
cd ~/personal/dotfiles
git add -A home/private_dot_claude/
git commit -m "refactor: remove chezmoi-managed claude config (replaced by claude-config repo)"
```

---

### Task 3: Update .chezmoiignore

**Files:**
- Modify: `home/.chezmoiignore` (lines 19-50)

**Step 1: Replace the .claude/ runtime exclusion block**

The current `.chezmoiignore` has ~22 lines listing individual `.claude/` runtime files (lines 19-41) plus the sandbox exclusion (lines 43-47). Since chezmoi no longer manages `~/.claude/` at all, replace the granular list with a single blanket ignore.

Replace lines 19-50 (from `# Claude Code runtime files` through `# Aurora host: deploy Claude config`) with:

```
# Claude Code config is managed by claude-config repo (symlinks), not chezmoi
.claude/
```

Keep the sandbox `setup-creds` exclusion but move it to a separate block:

```
# Sandbox: no credential seeding
{{- if eq .context "sandbox" }}
.local/bin/setup-creds
{{- end }}
```

**Step 2: Verify the ignore file**

Run: `cat ~/personal/dotfiles/home/.chezmoiignore`
Expected: `.claude/` is blanket-ignored for all environments. `setup-creds` excluded for sandbox only.

**Step 3: Commit**

```bash
cd ~/personal/dotfiles
git add home/.chezmoiignore
git commit -m "refactor: blanket-ignore .claude/ in chezmoiignore (managed by claude-config)"
```

---

### Task 4: Update bootstrap — claude-config clone + symlinks

**Files:**
- Modify: `home/run_once_before_bootstrap.sh.tmpl` (lines 24-31)

**Step 1: Replace old symlink cleanup with claude-config clone + symlink**

Replace lines 24-31 (the "Claude Code: replace old symlinks" block) with:

```bash
# ─── Claude Code: clone config repo and create symlinks ──────────────

{{ if ne .context "sandbox" -}}
CLAUDE_CONFIG_REPO="$HOME/personal/claude-config"
if [ ! -d "$CLAUDE_CONFIG_REPO" ]; then
    echo "Cloning claude-config..."
    mkdir -p "$HOME/personal"
    git clone https://github.com/rommelporras/claude-config.git "$CLAUDE_CONFIG_REPO"
fi

mkdir -p "$HOME/.claude"
for item in CLAUDE.md settings.json hooks skills agents; do
    if [ -e "$CLAUDE_CONFIG_REPO/$item" ]; then
        ln -sf "$CLAUDE_CONFIG_REPO/$item" "$HOME/.claude/$item"
    fi
done
echo "Linked ~/.claude/ → $CLAUDE_CONFIG_REPO"
{{ end -}}
```

Note: For distrobox, `$HOME` is container-local. The git clone URL works because the container has network access. But we actually want to symlink to the host's clone, not clone again inside the container. The `distrobox-setup.sh` script handles the container case separately (Task 7), so this bootstrap block only runs on Aurora and WSL where `$HOME` is the real home.

Wait — this bootstrap runs inside distrobox too (via `chezmoi init --apply`). We need to handle the distrobox case differently. Inside distrobox, claude-config lives on the host filesystem.

Revised approach for distrobox:

```bash
# ─── Claude Code: config repo symlinks ───────────────────────────────

{{ if ne .context "sandbox" -}}
# On Aurora/WSL: clone if needed. On Distrobox: host repo is already accessible.
if [ "$IS_DISTROBOX" = true ]; then
    # Normalize /var/home → /home for container compatibility
    HOST_HOME="/home/{{ .chezmoi.username }}"
    CLAUDE_CONFIG_REPO="$HOST_HOME/personal/claude-config"
else
    CLAUDE_CONFIG_REPO="$HOME/personal/claude-config"
    if [ ! -d "$CLAUDE_CONFIG_REPO" ]; then
        echo "Cloning claude-config..."
        mkdir -p "$HOME/personal"
        git clone https://github.com/rommelporras/claude-config.git "$CLAUDE_CONFIG_REPO"
    fi
fi

if [ -d "$CLAUDE_CONFIG_REPO" ]; then
    mkdir -p "$HOME/.claude"
    for item in CLAUDE.md settings.json hooks skills agents; do
        if [ -e "$CLAUDE_CONFIG_REPO/$item" ]; then
            ln -sf "$CLAUDE_CONFIG_REPO/$item" "$HOME/.claude/$item"
        fi
    done
    echo "Linked ~/.claude/ → $CLAUDE_CONFIG_REPO"
else
    echo "WARNING: claude-config not found at $CLAUDE_CONFIG_REPO — skipping Claude config"
fi
{{ end -}}
```

**Step 2: Verify template syntax**

Run: `cd ~/personal/dotfiles && chezmoi execute-template < home/run_once_before_bootstrap.sh.tmpl | head -40`

This may fail because chezmoi execute-template needs data. Instead, visually review the file.

**Step 3: Commit**

```bash
cd ~/personal/dotfiles
git add home/run_once_before_bootstrap.sh.tmpl
git commit -m "feat: bootstrap clones claude-config and creates symlinks"
```

---

### Task 5: Update bootstrap — Claude Code + NVM for Distrobox

**Files:**
- Modify: `home/run_once_before_bootstrap.sh.tmpl` (lines 79-84, 99-113)

**Step 1: Add Claude Code native installer**

After the Node.js block (line 84), add:

```bash
# Claude Code (WSL and Distrobox — Aurora has it via ujust bbrew)
{{ if ne .context "sandbox" -}}
if [ "$IS_AURORA" = false ] && ! command -v claude &>/dev/null; then
    echo "Installing Claude Code..."
    curl -fsSL https://claude.ai/install.sh | bash
fi
{{ end -}}
```

**Step 2: Expand NVM to include Distrobox**

Change the NVM section (lines 101-106) from:

```
{{ if eq .platform "wsl" -}}
# NVM (WSL only)
```

To:

```
{{ if or (eq .platform "wsl") (eq .platform "distrobox") -}}
# NVM (WSL and Distrobox — provides Node.js for Claude Code plugins)
```

Also change the closing `{{ end -}}` on line 113 — it currently wraps both NVM and Bun. Bun should remain WSL-only. Split them:

```
{{ if or (eq .platform "wsl") (eq .platform "distrobox") -}}
# NVM (WSL and Distrobox — provides Node.js for Claude Code plugins)
if [ ! -d "$HOME/.nvm" ]; then
    echo "Installing nvm..."
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash
fi
{{ end -}}

{{ if eq .platform "wsl" -}}
# Bun (WSL only)
if ! command -v bun &>/dev/null; then
    echo "Installing bun..."
    curl -fsSL https://bun.sh/install | bash
fi
{{ end -}}
```

**Step 3: Verify the template renders**

Visually review the file to ensure template blocks don't overlap and `{{ end -}}` tags match.

**Step 4: Commit**

```bash
cd ~/personal/dotfiles
git add home/run_once_before_bootstrap.sh.tmpl
git commit -m "feat: add Claude Code installer and NVM for distrobox"
```

---

### Task 6: Update setup-creds — plugin & MCP automation

**Files:**
- Modify: `home/dot_local/bin/executable_setup-creds.tmpl` (add before final "Done" echo)

**Step 1: Add plugin and MCP setup block**

Before the final `echo "=== Done ==="` (line 68-69), add:

```bash
# ─── Claude Code plugins & MCP ───────────────────────────────────────
if command -v claude &>/dev/null; then
    echo ""
    echo "--- Claude Code Plugins ---"

    # Marketplaces
    claude plugin marketplace add anthropics/claude-plugins-official 2>/dev/null || true
    claude plugin marketplace add obra/superpowers-marketplace 2>/dev/null || true

    # Plugins (user scope — available in all projects)
    claude plugin install context7@claude-plugins-official --scope user 2>/dev/null || true
    claude plugin install superpowers@superpowers-marketplace --scope user 2>/dev/null || true
    claude plugin install episodic-memory@superpowers-marketplace --scope user 2>/dev/null || true

    echo "Plugins installed"

    # Context7 MCP (needs 1Password API key)
    CONTEXT7_KEY=$($OP read 'op://Kubernetes/Context7/api-key' --no-newline 2>/dev/null)
    if [ -n "$CONTEXT7_KEY" ]; then
        claude mcp add --scope user --transport http context7 \
            https://mcp.context7.com/mcp \
            --header "CONTEXT7_API_KEY: $CONTEXT7_KEY"
        echo "Context7 MCP registered"
    else
        echo "WARNING: Could not read Context7 API key from 1Password"
    fi
else
    echo "Claude Code not installed — skipping plugin setup"
fi
```

Note: `$OP` is already defined at line 18 as `distrobox-host-exec op`. This variable is reused here.

**Step 2: Handle the distrobox-only guard**

Currently setup-creds has a guard at line 8-11 that exits if not in distrobox. The plugin setup should also work on Aurora/WSL where `op` is available directly. However, setup-creds is only deployed to distrobox containers (excluded for sandbox via .chezmoiignore, and on Aurora/WSL it's never called).

For now, keep the distrobox-only guard. On Aurora/WSL, plugin setup will remain a printed instruction in the bootstrap post-install section. This is acceptable because Aurora/WSL are set up once, manually.

Alternative: if we want setup-creds to work everywhere, remove the distrobox guard and make `$OP` conditional. But that changes the script's scope significantly. Keep it simple for now.

**Step 3: Commit**

```bash
cd ~/personal/dotfiles
git add home/dot_local/bin/executable_setup-creds.tmpl
git commit -m "feat: add Claude Code plugin and MCP setup to setup-creds"
```

---

### Task 7: Update distrobox-setup.sh

**Files:**
- Modify: `scripts/distrobox-setup.sh` (after line 81, before line 83)

**Step 1: Remove the separate claude-config symlink step**

The bootstrap (Task 4) now handles claude-config symlinks for distrobox too (using the host path). So `distrobox-setup.sh` doesn't need a separate symlink step — `chezmoi init --apply` on line 80 triggers the bootstrap which creates the symlinks.

No changes needed to distrobox-setup.sh for symlinks.

**Step 2: Verify the flow**

The existing flow in distrobox-setup.sh:
1. Install chezmoi (line 63-66)
2. Symlink chezmoi source to host repo (line 67-70)
3. Pre-seed chezmoi.toml (line 72-77)
4. `chezmoi init --apply` (line 80) → triggers bootstrap → clones/symlinks claude-config, installs NVM, installs Claude Code
5. `setup-creds` (line 87) → installs plugins, registers MCP, seeds credentials

This flow is correct. No code changes needed.

**Step 3: Commit**

No commit needed — no changes to this file.

---

### Task 8: Update AI sandbox Containerfile

**Files:**
- Modify: `containers/Containerfile.ai-sandbox` (line 28)

**Step 1: Replace npm install with native installer**

Change line 28 from:

```dockerfile
RUN npm install -g @anthropic-ai/claude-code
```

To:

```dockerfile
RUN curl -fsSL https://claude.ai/install.sh | bash
```

**Step 2: Verify Containerfile syntax**

Run: `cat ~/personal/dotfiles/containers/Containerfile.ai-sandbox`
Expected: native installer on the Claude Code line, Node.js still installed via NodeSource above.

**Step 3: Commit**

```bash
cd ~/personal/dotfiles
git add containers/Containerfile.ai-sandbox
git commit -m "refactor: use native Claude Code installer in AI sandbox"
```

---

### Task 9: Update bootstrap post-install instructions

**Files:**
- Modify: `home/run_once_before_bootstrap.sh.tmpl` (lines 204-210)

**Step 1: Update Claude Code setup instructions**

Replace lines 204-210 (the manual Claude Code plugin instructions) with:

```bash
{{ if ne .context "sandbox" -}}
if [ "$IS_DISTROBOX" = true ]; then
    echo "Claude Code plugins:"
    echo "  Run 'setup-creds' to install plugins and register Context7 MCP"
else
    echo "Claude Code plugins (run once):"
    echo "  claude plugin marketplace add anthropics/claude-plugins-official"
    echo "  claude plugin marketplace add obra/superpowers-marketplace"
    echo "  claude plugin install context7@claude-plugins-official --scope user"
    echo "  claude plugin install superpowers@superpowers-marketplace --scope user"
    echo "  claude plugin install episodic-memory@superpowers-marketplace --scope user"
    echo "  claude mcp add --scope user --transport http context7 https://mcp.context7.com/mcp \\"
    echo "    --header \"CONTEXT7_API_KEY: \$(op read 'op://Kubernetes/Context7/api-key' --no-newline)\""
fi
{{ end -}}
```

**Step 2: Commit**

```bash
cd ~/personal/dotfiles
git add home/run_once_before_bootstrap.sh.tmpl
git commit -m "docs: update post-install instructions for plugin setup"
```

---

### Task 10: Update project CLAUDE.md

**Files:**
- Modify: `CLAUDE.md` (root of dotfiles repo)

**Step 1: Update the Claude Code Config section**

Replace the "Claude Code Config" section (that references `home/private_dot_claude/`) with:

```markdown
## Claude Code Config

Claude Code global config (`~/.claude/`) is managed by a separate repo:
[claude-config](https://github.com/rommelporras/claude-config) — cloned to
`~/personal/claude-config` and symlinked into `~/.claude/`.

The bootstrap script handles cloning and symlinking. On distrobox containers,
symlinks point to the host's clone via absolute paths.

Plugin and MCP setup is handled by `setup-creds` (distrobox) or manual CLI
commands (Aurora/WSL) — see bootstrap post-install instructions.
```

**Step 2: Commit**

```bash
cd ~/personal/dotfiles
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md to reflect claude-config repo"
```

---

### Task 11: Test on Aurora host

**Step 1: Preview changes**

Run: `chezmoi diff`
Expected: shows bootstrap changes, .chezmoiignore changes. Should NOT show any .claude/ file changes (blanket-ignored now).

**Step 2: Apply**

Run: `chezmoi apply -v`
Expected: bootstrap runs, claude-config symlinks created.

**Step 3: Verify**

```bash
# Symlinks
ls -la ~/.claude/CLAUDE.md ~/.claude/settings.json ~/.claude/hooks ~/.claude/skills ~/.claude/agents

# Content resolves
head -3 ~/.claude/CLAUDE.md

# Claude Code
claude --version

# Node
node --version
```

Expected: all symlinks point to `~/personal/claude-config/`, Claude Code and Node available.

---

### Task 12: Test Distrobox personal

**Step 1: Destroy and recreate**

Run: `distrobox rm -f personal` (if exists)
Run: `scripts/distrobox-setup.sh personal`

**Step 2: Enter container and verify**

```bash
distrobox enter personal

# Symlinks
ls -la ~/.claude/CLAUDE.md ~/.claude/settings.json ~/.claude/hooks

# Content resolves (uses absolute host path)
head -3 ~/.claude/CLAUDE.md

# Claude Code
claude --version

# Node.js (via NVM)
source ~/.nvm/nvm.sh
node --version

# Plugins
claude plugin marketplace list
claude plugin list

# Context7 MCP
claude mcp list
```

Expected: symlinks to `/home/<user>/personal/claude-config/`, Claude Code installed, NVM + Node available, plugins installed, Context7 MCP registered.

---

### Task 13: Test Distrobox work-eam

**Step 1: Destroy and recreate**

Run: `distrobox rm -f work-eam` (if exists)
Run: `scripts/distrobox-setup.sh work-eam`

**Step 2: Enter and verify (same checks as Task 12)**

```bash
distrobox enter work-eam
ls -la ~/.claude/CLAUDE.md
claude --version
node --version
claude plugin list
claude mcp list
```

---

### Task 14: Test Distrobox sandbox

**Step 1: Destroy and recreate**

Run: `distrobox rm -f sandbox` (if exists)
Run: `scripts/distrobox-setup.sh sandbox`

**Step 2: Enter and verify NO claude config**

```bash
distrobox enter sandbox

# Should NOT exist
ls ~/.claude/ 2>&1
# Expected: No such file or directory (or empty dir)

# setup-creds should not exist
ls ~/.local/bin/setup-creds 2>&1
# Expected: No such file or directory
```

---

### Task 15: Test AI sandbox container build

**Step 1: Build**

Run: `podman build -t ai-sandbox-test -f containers/Containerfile.ai-sandbox containers/`

**Step 2: Verify Claude Code is installed**

Run: `podman run --rm ai-sandbox-test "claude --version"`
Expected: prints Claude Code version

**Step 3: Cleanup**

Run: `podman rmi ai-sandbox-test`
