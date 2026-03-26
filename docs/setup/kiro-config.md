# Kiro Config Setup

Kiro CLI and Kiro IDE configuration, managed via a public repo with symlinks into `~/.kiro/`.
Same pattern as `~/personal/claude-config` → `~/.claude/`.

## Scope

| Tool | Scope | Platform |
|---|---|---|
| **Kiro CLI** | `work-*` contexts only | WSL, distrobox (Aurora) |
| **Kiro IDE** | `work-*` contexts on WSL only | WSL (Windows app + WSL remoting) |
| **Claude Code** | `personal` + `work-*` contexts | All platforms (unchanged) |

## Repo structure

```
kiro-config/
├── .gitignore
├── README.md
├── steering/
│   ├── universal-rules.md
│   ├── environment.md
│   ├── engineering.md
│   └── tooling.md
├── agents/
│   └── default.json           # Global default agent (read-only tools pre-approved)
├── skills/                    # 14 skills total
│   ├── commit/SKILL.md
│   ├── push/SKILL.md
│   ├── explain-code/SKILL.md
│   ├── systematic-debugging/SKILL.md
│   ├── verification-before-completion/SKILL.md
│   ├── writing-plans/SKILL.md
│   ├── test-driven-development/SKILL.md
│   ├── requesting-code-review/SKILL.md
│   ├── executing-plans/SKILL.md
│   ├── brainstorming/SKILL.md
│   ├── dispatching-parallel-agents/SKILL.md
│   ├── subagent-driven-development/SKILL.md
│   ├── receiving-code-review/SKILL.md
│   └── finishing-a-development-branch/SKILL.md
└── settings/
    ├── cli.json
    └── mcp.json               # 9 AWS MCP servers
```

## Symlink setup

```bash
# Back up Kiro defaults
mv ~/.kiro/agents ~/.kiro/agents.bak 2>/dev/null
mv ~/.kiro/settings ~/.kiro/settings.bak 2>/dev/null

# Create symlinks
ln -sfn ~/personal/kiro-config/steering ~/.kiro/steering
ln -sfn ~/personal/kiro-config/agents ~/.kiro/agents
ln -sfn ~/personal/kiro-config/skills ~/.kiro/skills
ln -sfn ~/personal/kiro-config/settings ~/.kiro/settings
```

## Chezmoi integration

Already applied:
- `home/.chezmoiignore` — `.kiro/` ignored (same as `.claude/`)
- `home/dot_zshrc.tmpl` — Kiro CLI pre/post blocks, `kiro()` IDE function, auto-sync pull
- `scripts/wsl_setup.py` — clones kiro-config for `work-*` contexts, creates symlinks

## CLI settings (set manually after install)

```bash
kiro-cli settings chat.defaultModel auto
kiro-cli settings chat.enableCheckpoint true
kiro-cli settings chat.defaultAgent default
```

These are stored in `~/.kiro/settings/cli.json` but are machine-local (not in the repo)
because they include per-machine state like `mcp.loadedBefore`.

## Global agent (default.json)

Pre-approves read-only tools and sets trusted paths:
- `allowedTools`: read, grep, glob, web_fetch
- `toolsSettings.fs_read.allowedPaths`: ~/.kiro, ~/personal, ~/eam
- `toolsSettings.fs_write.allowedPaths`: ~/personal, ~/eam
- `toolsSettings.execute_bash.autoAllowReadonly`: true
- `toolsSettings.use_aws.autoAllowReadonly`: true

SRE agent is **project-level** (in each project's `.kiro/agents/`), not global.

## MCP servers (mcp.json)

9 AWS MCP servers, no hardcoded profiles or regions.
Set `AWS_PROFILE` before launching Kiro CLI.

| Server | Package |
|---|---|
| AWS Documentation | `awslabs.aws-documentation-mcp-server` |
| ECS | `awslabs.ecs-mcp-server` (uses `--from` flag) |
| EKS | `awslabs.eks-mcp-server` |
| Terraform | `awslabs.terraform-mcp-server` |
| CloudWatch | `awslabs.cloudwatch-mcp-server` |
| CloudTrail | `awslabs.cloudtrail-mcp-server` |
| IAM | `awslabs.iam-mcp-server` |
| Diagram | `awslabs.aws-diagram-mcp-server` |
| Cost Explorer | `awslabs.cost-explorer-mcp-server` |

## Skills

3 original (commit, push, explain-code) + 11 ported from superpowers.
All use Kiro-compatible frontmatter (name + description only).
No Claude-specific references.

## Kiro IDE on WSL

See `~/eam/eam-sre/rommel-porras/docs/kiro-ide-wsl-setup.md`:
1. `argv.json` on Windows: enable proposed API for `jeanp413.open-remote-wsl`
2. Install Open Remote - WSL extension in Kiro
3. `kiro()` shell function in `dot_zshrc.tmpl` (WSL + `work-*` only, auto-detects install path)

## Kiro on Aurora (distrobox) — future

- **Kiro CLI in distrobox**: install via curl script inside `work-*` distrobox container
- **Kiro IDE via distrobox**: possible with `distrobox-export`, needs testing

## Checklist

- [x] Create `~/personal/kiro-config` repo locally
- [x] Write steering files (translated from CLAUDE.md)
- [x] Adapt commit/push/explain-code skills
- [x] Port 11 superpowers skills
- [x] Create default agent JSON with toolsSettings
- [x] Configure 9 MCP servers in mcp.json (correct package names)
- [x] Create symlinks into ~/.kiro/
- [x] Update `scripts/wsl_setup.py` to clone kiro-config + create symlinks
- [x] Add kiro-config auto-sync to dot_zshrc.tmpl
- [x] Add `.kiro/` to `.chezmoiignore`
- [x] Set CLI settings (defaultModel, enableCheckpoint, defaultAgent)
- [x] Test on WSL work-eam context — all 14 skills load, 9 MCP servers load
- [ ] Create GitHub repo (public) and push
- [ ] Commit dotfiles changes
- [ ] Test on Aurora distrobox (future)
